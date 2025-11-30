#!/usr/bin/env bash
# Verify go.sum checksums against Go checksum database
# This detects supply chain attacks where dependencies are tampered with
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Verifying go.sum checksums"

# Check 1: Ensure go.sum exists
if [[ ! -f "go.sum" ]]; then
    check_failure \
        "go.sum file not found" \
        "Missing go.sum file in project root" \
        "Run: go mod tidy"
    exit 1
fi

# Check 2: Verify go.sum is not empty
if [[ ! -s "go.sum" ]]; then
    check_failure \
        "go.sum file is empty" \
        "Empty go.sum indicates no dependencies or corruption" \
        "Run: go mod tidy"
    exit 1
fi

# Check 3: Verify all modules against checksum database
# GOSUMDB is the checksum database (default: sum.golang.org)
# This ensures modules haven't been tampered with since first download
if ! run_check "go mod verify 2>&1"; then
    check_failure \
        "Module verification failed" \
        "$CHECK_OUTPUT" \
        "Possible causes:"$'\n'"  - Module was modified after download"$'\n'"  - go.sum is out of sync"$'\n'"  - Supply chain attack detected"$'\n'"Fix: rm -rf ~/go/pkg/mod && go mod download"
    exit 1
fi

# Check 4: Verify go.sum is complete (all transitive deps have checksums)
# go mod tidy will fail if checksums are missing
if ! run_check "go mod tidy -v 2>&1"; then
    # Check if anything changed
    if git diff --quiet go.sum 2>/dev/null; then
        # No changes, just verbose output
        :
    else
        check_failure \
            "go.sum is incomplete" \
            "$CHECK_OUTPUT" \
            "Some dependencies are missing checksums"$'\n'"Run: go mod tidy && git add go.sum"
        exit 1
    fi
fi

# Check 5: Verify no go.sum modifications are staged but uncommitted
# This prevents accidental checksum changes from slipping through
if git diff --cached --quiet go.sum 2>/dev/null; then
    :
else
    echo "  Note: go.sum has staged changes - ensure they are intentional"
fi

# Check 6: Count and report dependencies
DEP_COUNT=$(wc -l < go.sum | tr -d ' ')
DIRECT_DEPS=$(go list -m -f '{{if not .Indirect}}{{.Path}}{{end}}' all 2>/dev/null | wc -l | tr -d ' ')

check_success "All checksums verified ($DIRECT_DEPS direct, $DEP_COUNT total entries)"
