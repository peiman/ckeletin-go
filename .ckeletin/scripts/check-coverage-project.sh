#!/bin/bash
# Check if project coverage meets minimum threshold
# Similar to codecov/project check

set -e

COVERAGE_FILE="${COVERAGE_FILE:-coverage.txt}"
MIN_COVERAGE="${MIN_COVERAGE:-85.0}"

if [ ! -f "$COVERAGE_FILE" ]; then
    echo "❌ Coverage file not found: $COVERAGE_FILE"
    echo "Run 'task test' first to generate coverage data"
    exit 1
fi

# Calculate total coverage using go tool cover
total_coverage=$(go tool cover -func="$COVERAGE_FILE" | grep "total:" | awk '{print $3}' | sed 's/%//')

if [ -z "$total_coverage" ]; then
    echo "❌ Failed to parse coverage data"
    exit 1
fi

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
