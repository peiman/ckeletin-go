#!/usr/bin/env bash
# Run all linters with standardized output
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Linting code"

# Run go vet first
if ! run_check "go vet ./... 2>&1"; then
    check_failure \
        "go vet found issues" \
        "$CHECK_OUTPUT" \
        "Fix the issues reported by go vet"
    exit 1
fi

# Run golangci-lint
if ! run_check "golangci-lint run 2>&1"; then
    check_failure \
        "Linting failed" \
        "$CHECK_OUTPUT" \
        "Fix the linting issues reported above"
    exit 1
fi

check_success "No linting issues found"
