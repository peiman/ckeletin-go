#!/bin/bash
# Check if patch/diff coverage meets minimum threshold.
# For each changed line, checks whether any coverage block containing
# that line was exercised by tests.

set -eo pipefail

COVERAGE_FILE="${COVERAGE_FILE:-coverage.txt}"
MIN_PATCH_COVERAGE="${MIN_PATCH_COVERAGE:-80.0}"
BASE_BRANCH="${BASE_BRANCH:-main}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Coverage file not found: $COVERAGE_FILE"
    echo "Run 'task test' first to generate coverage data"
    exit 1
fi

# Get list of changed .go files (excluding test files and non-production code)
# Exclusions:
#   _test.go      — test files don't need coverage
#   scripts/      — top-level build scripts
#   .ckeletin/scripts/ — framework build-time scripts (standalone go run, own test suite)
#   internal/testutil/ — test utilities (platform-specific, can't achieve full coverage)
#   /demo/        — demo code for documentation
#   _tui.go       — TUI code requiring interactive testing
if git rev-parse --verify "$BASE_BRANCH" &>/dev/null; then
    changed_files=$(git diff "$BASE_BRANCH"...HEAD --name-only --diff-filter=AM | grep '\.go$' | grep -v '_test\.go$' | grep -v '^scripts/' | grep -v '^\.ckeletin/scripts/' | grep -v '^internal/testutil/' | grep -v '/demo/' | grep -v '_tui\.go$' || true)
else
    # Fallback to staged changes
    changed_files=$(git diff --cached --name-only --diff-filter=AM | grep '\.go$' | grep -v '_test\.go$' | grep -v '^scripts/' | grep -v '^\.ckeletin/scripts/' | grep -v '^internal/testutil/' | grep -v '/demo/' | grep -v '_tui\.go$' || true)
fi

if [ -z "$changed_files" ]; then
    echo "ℹ️  No Go files changed - patch coverage check skipped"
    exit 0
fi

echo "📝 Changed files:"
echo "$changed_files" | sed 's/^/  - /'
echo ""

# get_changed_lines extracts added/modified line numbers from git diff for a file.
get_changed_lines() {
    local file="$1"
    if git rev-parse --verify "$BASE_BRANCH" &>/dev/null; then
        git diff "$BASE_BRANCH"...HEAD --unified=0 -- "$file"
    else
        git diff --cached --unified=0 -- "$file"
    fi | grep '^@@' | sed -E 's/^@@ -[0-9,]+ \+([0-9]+)(,([0-9]+))? @@.*/\1 \3/' | while read -r start count; do
        count=${count:-1}
        local end=$((start + count - 1))
        for ((i = start; i <= end; i++)); do
            echo "$i"
        done
    done | sort -n | uniq
}

# is_line_covered checks if a line number falls within ANY coverage block that has hits > 0.
# Reads coverage entries from the file at cov_file.
is_line_covered() {
    local line_num="$1"
    local cov_file="$2"

    while IFS= read -r entry; do
        # Parse: file.go:startLine.startCol,endLine.endCol numStmts numHits
        if [[ $entry =~ :([0-9]+)\.[0-9]+,([0-9]+)\.[0-9]+[[:space:]]+[0-9]+[[:space:]]+([0-9]+)$ ]]; then
            local block_start="${BASH_REMATCH[1]}"
            local block_end="${BASH_REMATCH[2]}"
            local hits="${BASH_REMATCH[3]}"

            if [ "$line_num" -ge "$block_start" ] && [ "$line_num" -le "$block_end" ] && [ "$hits" -gt 0 ]; then
                return 0  # covered
            fi
        fi
    done < "$cov_file"
    return 1  # not covered
}

# Main loop: for each changed file, check coverage of each changed line
cov_tmpdir=$(mktemp -d) || { echo "ERROR: Failed to create temp directory" >&2; exit 1; }
trap 'rm -rf "$cov_tmpdir"' EXIT

total_lines=0
covered_lines=0
file_idx=0

while IFS= read -r file; do
    [ -f "$file" ] || continue

    # Get all coverage entries for this file (may contain overlapping blocks — that's fine,
    # we only need ANY block with hits > 0 to mark a line as covered)
    file_idx=$((file_idx + 1))
    cov_tmp="$cov_tmpdir/cov_$file_idx"
    grep "$(basename "$file")" "$COVERAGE_FILE" | grep "/$file:" > "$cov_tmp" 2>/dev/null || true

    if [ ! -s "$cov_tmp" ]; then
        echo "⚠️  No coverage data for $file"
        continue
    fi

    # Get changed lines
    changed_lines=$(get_changed_lines "$file")
    if [ -z "$changed_lines" ]; then
        continue
    fi

    file_total=0
    file_covered=0

    while read -r line_num; do
        # Check if this line falls within ANY coverage block (even uncovered ones)
        # If it doesn't fall in any block, it's a non-statement line (comment, blank, etc.)
        in_any_block=false
        while IFS= read -r entry; do
            if [[ $entry =~ :([0-9]+)\.[0-9]+,([0-9]+)\.[0-9]+[[:space:]] ]]; then
                local_start="${BASH_REMATCH[1]}"
                local_end="${BASH_REMATCH[2]}"
                if [ "$line_num" -ge "$local_start" ] && [ "$line_num" -le "$local_end" ]; then
                    in_any_block=true
                    break
                fi
            fi
        done < "$cov_tmp"

        if ! $in_any_block; then
            continue  # skip non-statement lines
        fi

        file_total=$((file_total + 1))
        if is_line_covered "$line_num" "$cov_tmp"; then
            file_covered=$((file_covered + 1))
        fi
    done <<< "$changed_lines"

    if [ "$file_total" -gt 0 ]; then
        file_pct=$(echo "scale=1; $file_covered * 100 / $file_total" | bc -l)
        echo "  $file: ${file_pct}% (${file_covered}/${file_total} changed lines)"
        total_lines=$((total_lines + file_total))
        covered_lines=$((covered_lines + file_covered))
    fi
done <<< "$changed_files"

echo ""

if [ "$total_lines" -eq 0 ]; then
    echo "ℹ️  No measurable statements in changed files"
    exit 0
fi

# Calculate patch coverage percentage
patch_coverage=$(echo "scale=2; $covered_lines * 100 / $total_lines" | bc -l)

echo "📊 Patch Coverage: ${patch_coverage}% (${covered_lines}/${total_lines} changed lines)"
echo "🎯 Minimum Required: ${MIN_PATCH_COVERAGE}%"

# Compare coverage
result=$(echo "$patch_coverage >= $MIN_PATCH_COVERAGE" | bc -l)

if [ "$result" -eq 1 ]; then
    echo "✅ Patch coverage check passed!"
    exit 0
else
    diff=$(echo "$MIN_PATCH_COVERAGE - $patch_coverage" | bc -l | xargs printf "%.2f")
    echo "❌ Patch coverage check failed!"
    echo "   Need ${diff}% more coverage on changed files"
    echo ""
    echo "💡 Tips:"
    echo "   - Add tests for new functions"
    echo "   - Cover error paths and edge cases"
    echo "   - Run 'task test:coverage:html' to see uncovered lines"
    exit 1
fi
