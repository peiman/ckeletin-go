#!/bin/bash
# Single source of truth for Go formatting
# Usage: ./scripts/format-go.sh [fix|check] [files...]
#
# Modes:
#   fix   - Format files in place (default)
#   check - Check if files are formatted, fail if not (CI mode)

set -eo pipefail

# Source standard output functions and the scanner-exclusion SSOT
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

MODE="${1:-fix}"
shift || true
FILES=("$@")
if [ ${#FILES[@]} -eq 0 ]; then
    # goimports/gofmt given "." descend into EVERYTHING, including other
    # worktrees' checkouts under .claude/worktrees/ — enumerate the tree
    # ourselves with the shared exclusions instead
    while IFS= read -r f; do
        FILES+=("$f")
    done < <(ckeletin_go_files .)
    if [ ${#FILES[@]} -eq 0 ]; then
        echo "No Go files to format"
        exit 0
    fi
fi

format_files() {
    goimports -w "${FILES[@]}"
    gofmt -s -w "${FILES[@]}"
}

check_files() {
    check_header "Checking code formatting"

    # A missing formatter must fail the gate loudly — swallowing it would
    # turn the check into a silent pass (fail-open)
    local tool
    for tool in goimports gofmt; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            check_failure \
                "Formatting check failed" \
                "Required formatter '$tool' not found in PATH" \
                "Run: task setup"
            exit 1
        fi
    done

    # stderr is captured separately so tool errors never masquerade as
    # "files needing formatting", and a crashing formatter fails the gate
    local err_file
    err_file=$(mktemp)
    trap 'rm -f "$err_file"' EXIT

    local unformatted_output=""

    # Check goimports
    local goimports_output
    if ! goimports_output=$(goimports -l "${FILES[@]}" 2>"$err_file"); then
        check_failure \
            "goimports failed to run" \
            "$(cat "$err_file")" \
            "Fix the goimports error above; if the tool is broken, run: task setup"
        exit 1
    fi
    if [ -n "$goimports_output" ]; then
        unformatted_output+="Files need goimports:"$'\n'"$goimports_output"$'\n\n'
    fi

    # Check gofmt
    local gofmt_output
    if ! gofmt_output=$(gofmt -l "${FILES[@]}" 2>"$err_file"); then
        check_failure \
            "gofmt failed to run" \
            "$(cat "$err_file")" \
            "Fix the gofmt error above; if the tool is broken, run: task setup"
        exit 1
    fi
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
