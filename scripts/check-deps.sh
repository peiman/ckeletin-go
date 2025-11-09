#!/usr/bin/env bash
# Check dependency integrity and vulnerabilities
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Checking dependency integrity"

# Check 1: Verify dependencies haven't been modified
if ! run_check "go mod verify 2>&1"; then
    check_failure \
        "Dependency verification failed" \
        "$CHECK_OUTPUT" \
        "Run: go mod tidy"
    exit 1
fi

# Check 2: Check for outdated dependencies (non-blocking, just informational)
OUTDATED_OUTPUT=$(go list -u -m -json all 2>/dev/null | go-mod-outdated -update -direct 2>&1 || true)

# Check 3: Check for vulnerabilities
if ! run_check "govulncheck ./... 2>&1"; then
    check_failure \
        "Security vulnerabilities found" \
        "$CHECK_OUTPUT" \
        "Review vulnerabilities and update dependencies"
    exit 1
fi

check_success "All dependencies verified and secure"
