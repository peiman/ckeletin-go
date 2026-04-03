#!/bin/bash
# Check if project coverage meets minimum threshold
# Similar to codecov/project check
#
# Excludes from coverage calculation:
# - _tui.go files (TUI code requires interactive testing)
# - /demo/ directories (demo code is for documentation)

set -eo pipefail

COVERAGE_FILE="${COVERAGE_FILE:-coverage.txt}"
MIN_COVERAGE="${MIN_COVERAGE:-85.0}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Coverage file not found: $COVERAGE_FILE"
    echo "Run 'task test' first to generate coverage data"
    exit 1
fi

# Calculate coverage ourselves, excluding TUI and demo code
# Format: file:line.col,line.col numStatements numHits
#
# When using -coverpkg=./..., each test binary emits coverage data for ALL
# packages. gocovmerge can produce duplicate entries for the same block with
# different hit counts (e.g., 0 from binaries that don't touch the code, >0
# from the package's own tests). We deduplicate by taking the maximum hit
# count per unique block before calculating coverage.
#
# Exclusions:
#   - *_tui.go: Legacy TUI file naming convention
#   - internal/check/executor.go: TUI executor (split from check_tui.go)
#   - internal/check/summary.go: TUI summary rendering (split from check_tui.go)
#   - /demo/: Demo code for documentation

# Calculate per-package coverage from go test output.
#
# The primary coverage file (coverage.txt) uses -coverpkg=./... which instruments
# each source file once per test binary. This creates overlapping coverage blocks
# with different boundaries that inflate the total statement count and undercount
# actual coverage. For accurate threshold enforcement, we generate a separate
# per-package coverage profile where each package is only instrumented once.
PERPKG_COV=$(mktemp)
trap 'rm -f "$PERPKG_COV"' EXIT

go test -tags dev -coverprofile="$PERPKG_COV" -covermode=atomic ./... ./.ckeletin/pkg/... 2>/dev/null

if [ ! -s "$PERPKG_COV" ]; then
    echo "❌ Failed to generate per-package coverage profile"
    echo "Falling back to primary coverage file"
    PERPKG_COV="$COVERAGE_FILE"
fi

total_statements=0
covered_statements=0

while IFS= read -r line; do
    # Skip mode line, TUI, and demo code
    if [[ "$line" == mode:* ]] || [[ "$line" == *"_tui.go"* ]] || [[ "$line" == *"/demo/"* ]] \
       || [[ "$line" == *"internal/check/executor.go"* ]] \
       || [[ "$line" == *"internal/check/summary.go"* ]]; then
        continue
    fi

    if [[ $line =~ ([0-9]+)[[:space:]]+([0-9]+)$ ]]; then
        stmts="${BASH_REMATCH[1]}"
        hits="${BASH_REMATCH[2]}"

        total_statements=$((total_statements + stmts))
        if [ "$hits" -gt 0 ]; then
            covered_statements=$((covered_statements + stmts))
        fi
    fi
done < "$PERPKG_COV"

if [ "$total_statements" -eq 0 ]; then
    echo "❌ Failed to parse coverage data"
    exit 1
fi

total_coverage=$(echo "scale=1; $covered_statements * 100 / $total_statements" | bc -l)

# Compare coverage (using bc for floating point comparison)
if command -v bc &> /dev/null; then
    result=$(echo "$total_coverage >= $MIN_COVERAGE" | bc -l)
else
    # Fallback to awk if bc not available
    result=$(awk -v tc="$total_coverage" -v min="$MIN_COVERAGE" 'BEGIN {print (tc >= min)}')
fi

echo "📊 Project Coverage: ${total_coverage}%"
echo "🎯 Minimum Required: ${MIN_COVERAGE}%"

if [ "$result" -eq 1 ]; then
    echo "✅ Coverage check passed!"
    exit 0
else
    diff=$(awk -v tc="$total_coverage" -v min="$MIN_COVERAGE" 'BEGIN {printf "%.2f", min - tc}')
    echo "❌ Coverage check failed!"
    echo "   Need ${diff}% more coverage to reach ${MIN_COVERAGE}%"
    exit 1
fi
