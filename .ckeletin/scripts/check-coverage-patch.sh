#!/bin/bash
# Check if patch/diff coverage meets minimum threshold
# True patch coverage: only counts coverage blocks that overlap with changed lines

set -eo pipefail

COVERAGE_FILE="${COVERAGE_FILE:-coverage.txt}"
MIN_PATCH_COVERAGE="${MIN_PATCH_COVERAGE:-80.0}"
BASE_BRANCH="${BASE_BRANCH:-main}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Coverage file not found: $COVERAGE_FILE"
    echo "Run 'task test' first to generate coverage data"
    exit 1
fi

# Get list of changed .go files (excluding _test.go, scripts/, .ckeletin/scripts/, testutil/, demo/, and _tui.go)
# testutil is excluded because platform-specific skip helpers can't achieve 100% coverage on any single platform
# demo is excluded because demo code is meant for documentation, not production
# _tui.go files are excluded because TUI code requires interactive testing that's difficult to unit test
# .ckeletin/scripts/ is excluded because these are standalone build-time scripts (run via go run)
#   with their own test suites, not compiled into the binary — they don't appear in coverage.txt
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
# Returns one line number per line, sorted and deduplicated.
get_changed_lines() {
    local file="$1"
    local diff_cmd

    if git rev-parse --verify "$BASE_BRANCH" &>/dev/null; then
        diff_cmd="git diff $BASE_BRANCH...HEAD --unified=0 -- $file"
    else
        diff_cmd="git diff --cached --unified=0 -- $file"
    fi

    # Parse @@ hunk headers: @@ -old,count +new,count @@
    # Extract the +new,count part (lines added/modified in the new version)
    $diff_cmd | grep '^@@' | sed -E 's/^@@ -[0-9,]+ \+([0-9]+)(,([0-9]+))? @@.*/\1 \3/' | while read -r start count; do
        count=${count:-1}
        local end=$((start + count - 1))
        for ((i = start; i <= end; i++)); do
            echo "$i"
        done
    done | sort -n | uniq
}

# Check if a coverage block (startLine-endLine) overlaps with any changed line
block_overlaps_changes() {
    local block_start="$1"
    local block_end="$2"
    local changed_lines_file="$3"

    while read -r changed_line; do
        if [ "$changed_line" -ge "$block_start" ] && [ "$changed_line" -le "$block_end" ]; then
            return 0  # overlap found
        fi
    done < "$changed_lines_file"
    return 1  # no overlap
}

# Parse coverage for changed files — only count blocks that overlap with changed lines
# Coverage format: github.com/user/repo/file.go:startLine.startCol,endLine.endCol numStmts numHits
total_statements=0
covered_statements=0

while IFS= read -r file; do
    # Skip if file doesn't exist
    [ -f "$file" ] || continue

    # Get coverage data for this file
    file_data=$(grep "$(basename "$file")" "$COVERAGE_FILE" | grep "/$file:" || true)

    if [ -z "$file_data" ]; then
        echo "⚠️  No coverage data for $file"
        continue
    fi

    # Get changed line numbers for this file
    changed_lines_tmp=$(mktemp)
    get_changed_lines "$file" > "$changed_lines_tmp"

    if [ ! -s "$changed_lines_tmp" ]; then
        rm -f "$changed_lines_tmp"
        continue
    fi

    file_total=0
    file_covered=0

    while IFS= read -r line; do
        # Parse: file.go:10.2,12.3 5 2
        # Full regex: capture startLine, endLine, numStmts, numHits
        if [[ $line =~ :([0-9]+)\.[0-9]+,([0-9]+)\.[0-9]+[[:space:]]+([0-9]+)[[:space:]]+([0-9]+)$ ]]; then
            block_start="${BASH_REMATCH[1]}"
            block_end="${BASH_REMATCH[2]}"
            stmts="${BASH_REMATCH[3]}"
            hits="${BASH_REMATCH[4]}"

            # Only count this block if it overlaps with changed lines
            if block_overlaps_changes "$block_start" "$block_end" "$changed_lines_tmp"; then
                file_total=$((file_total + stmts))
                if [ "$hits" -gt 0 ]; then
                    file_covered=$((file_covered + stmts))
                fi
            fi
        fi
    done <<< "$file_data"

    rm -f "$changed_lines_tmp"

    if [ "$file_total" -gt 0 ]; then
        file_pct=$(echo "scale=1; $file_covered * 100 / $file_total" | bc -l)
        echo "  $file: ${file_pct}% (${file_covered}/${file_total} statements)"
        total_statements=$((total_statements + file_total))
        covered_statements=$((covered_statements + file_covered))
    fi
done <<< "$changed_files"

echo ""

if [ "$total_statements" -eq 0 ]; then
    echo "ℹ️  No measurable statements in changed files"
    exit 0
fi

# Calculate patch coverage percentage
patch_coverage=$(echo "scale=2; $covered_statements * 100 / $total_statements" | bc -l)

echo "📊 Patch Coverage: ${patch_coverage}% (${covered_statements}/${total_statements} statements)"
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
