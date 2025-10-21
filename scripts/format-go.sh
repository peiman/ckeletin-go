#!/bin/bash
# Single source of truth for Go formatting
# Usage: ./scripts/format-go.sh [fix|check] [files...]
#
# Modes:
#   fix   - Format files in place (default)
#   check - Check if files are formatted, fail if not (CI mode)

set -e

MODE="${1:-fix}"
shift || true
FILES="${@:-.}"

format_files() {
    goimports -w $FILES
    gofmt -s -w $FILES
}

check_files() {
    local needs_format=0

    # Check goimports
    unformatted=$(goimports -l $FILES 2>/dev/null || true)
    if [ -n "$unformatted" ]; then
        echo "❌ Files need goimports formatting:"
        echo "$unformatted"
        needs_format=1
    fi

    # Check gofmt
    unformatted=$(gofmt -l $FILES 2>/dev/null || true)
    if [ -n "$unformatted" ]; then
        echo "❌ Files need gofmt formatting:"
        echo "$unformatted"
        needs_format=1
    fi

    if [ $needs_format -eq 1 ]; then
        echo ""
        echo "💡 Run 'task format' to fix formatting issues"
        exit 1
    fi

    echo "✅ All Go files properly formatted"
}

case "$MODE" in
    check)
        check_files
        ;;
    fix)
        format_files
        ;;
    *)
        echo "Unknown mode: $MODE"
        echo "Usage: $0 [fix|check] [files...]"
        exit 1
        ;;
esac
