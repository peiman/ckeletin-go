#!/bin/bash
# Check if project coverage meets minimum threshold
# Similar to codecov/project check
#
# Excludes from coverage calculation:
# - _tui.go files (TUI code requires interactive testing)
# - /demo/ directories (demo code is for documentation)

set -e

COVERAGE_FILE="${COVERAGE_FILE:-coverage.txt}"
MIN_COVERAGE="${MIN_COVERAGE:-85.0}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Coverage file not found: $COVERAGE_FILE"
    echo "Run 'task test' first to generate coverage data"
    exit 1
fi

# Calculate coverage ourselves, excluding TUI and demo code
# Format: file:line.col,line.col numStatements numHits
total_statements=0
covered_statements=0

while IFS= read -r line; do
    # Skip lines containing _tui.go or /demo/
    if [[ "$line" == *"_tui.go"* ]] || [[ "$line" == *"/demo/"* ]]; then
        continue
    fi

    # Parse: file:10.2,12.3 5 2 where 5 is statements, 2 is hits
    if [[ $line =~ ([0-9]+)[[:space:]]+([0-9]+)$ ]]; then
        stmts="${BASH_REMATCH[1]}"
        hits="${BASH_REMATCH[2]}"

        total_statements=$((total_statements + stmts))
        if [ "$hits" -gt 0 ]; then
            covered_statements=$((covered_statements + stmts))
        fi
    fi
done < "$COVERAGE_FILE"

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
