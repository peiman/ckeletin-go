#!/bin/bash
# Single source of truth for Go formatting
# Usage: ./scripts/format-go.sh [fix|check] [files...]
#
# Modes:
#   fix   - Format files in place (default)
#   check - Check if files are formatted, fail if not (CI mode)

set -e

# Source standard output functions (only for check mode)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

MODE="${1:-fix}"
shift || true
FILES="${@:-.}"

format_files() {
    goimports -w $FILES
    gofmt -s -w $FILES
}

check_files() {
    # shellcheck source=lib/check-output.sh
    source "${SCRIPT_DIR}/lib/check-output.sh"

    check_header "Checking code formatting"

    local unformatted_output=""

    # Check goimports
    local goimports_output=$(goimports -l $FILES 2>/dev/null || true)
    if [ -n "$goimports_output" ]; then
        unformatted_output+="Files need goimports:"$'\n'"$goimports_output"$'\n\n'
    fi

    # Check gofmt
    local gofmt_output=$(gofmt -l $FILES 2>/dev/null || true)
    if [ -n "$gofmt_output" ]; then
        unformatted_output+="Files need gofmt:"$'\n'"$gofmt_output"
    fi

    if [ -n "$unformatted_output" ]; then
        check_failure \
            "Formatting check failed" \
            "$unformatted_output" \
            "Run: task format"
        exit 1
    fi

    check_success "All Go files properly formatted"
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
