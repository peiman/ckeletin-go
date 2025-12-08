#!/bin/bash
# Check if patch/diff coverage meets minimum threshold
# Similar to codecov/patch check - only checks coverage of changed lines

set -e

COVERAGE_FILE="${COVERAGE_FILE:-coverage.txt}"
MIN_PATCH_COVERAGE="${MIN_PATCH_COVERAGE:-80.0}"
BASE_BRANCH="${BASE_BRANCH:-main}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Coverage file not found: $COVERAGE_FILE"
    echo "Run 'task test' first to generate coverage data"
    exit 1
fi

# Get list of changed .go files (excluding _test.go, scripts/, and testutil/)
# testutil is excluded because platform-specific skip helpers can't achieve 100% coverage on any single platform
if git rev-parse --verify "$BASE_BRANCH" &>/dev/null; then
    changed_files=$(git diff "$BASE_BRANCH"...HEAD --name-only --diff-filter=AM | grep '\.go$' | grep -v '_test\.go$' | grep -v '^scripts/' | grep -v '^internal/testutil/' || true)
else
    # Fallback to staged changes
    changed_files=$(git diff --cached --name-only --diff-filter=AM | grep '\.go$' | grep -v '_test\.go$' | grep -v '^scripts/' | grep -v '^internal/testutil/' || true)
fi

if [ -z "$changed_files" ]; then
    echo "ℹ️  No Go files changed - patch coverage check skipped"
    exit 0
fi

echo "📝 Changed files:"
echo "$changed_files" | sed 's/^/  - /'
echo ""

# Parse coverage for changed files
# Format: github.com/user/repo/file.go:startLine.startCol,endLine.endCol numStmts numHits
total_statements=0
covered_statements=0

while IFS= read -r file; do
    # Skip if file doesn't exist
    [ -f "$file" ] || continue
    
    # Count statements and coverage for this file
    # The coverage file has format: full/path/to/file.go:lines numStatements execCount
    file_data=$(grep "$(basename "$file")" "$COVERAGE_FILE" | grep "/$file:" || true)
    
    if [ -z "$file_data" ]; then
        echo "⚠️  No coverage data for $file"
        continue
    fi
    
    file_total=0
    file_covered=0
    
    while IFS= read -r line; do
        # Parse: file.go:10.2,12.3 5 2
        # Where 5 is number of statements, 2 is execution count
        if [[ $line =~ ([0-9]+)[[:space:]]+([0-9]+)$ ]]; then
            stmts="${BASH_REMATCH[1]}"
            hits="${BASH_REMATCH[2]}"
            
            file_total=$((file_total + stmts))
            if [ "$hits" -gt 0 ]; then
                file_covered=$((file_covered + stmts))
            fi
        fi
    done <<< "$file_data"
    
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
    echo "   - Run 'task test:coverage-html' to see uncovered lines"
    exit 1
fi
