#!/usr/bin/env bash
# Check dependency licenses using go-licenses (source-based, fast)
# Uses conservative permissive-only policy by default
# See: ADR-011 License Compliance Strategy

set -e

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

# Default policy: Allow permissive licenses only
# Note: go-licenses doesn't support both --allowed_licenses and --disallowed_types
# We use --allowed_licenses for explicit permissive-only policy
ALLOWED_LICENSES="${LICENSE_ALLOWED:-MIT,Apache-2.0,BSD-2-Clause,BSD-3-Clause,ISC,0BSD,Unlicense}"

# Get module path to ignore self
MODULE_PATH=$(go list -m 2>/dev/null || echo "github.com/peiman/ckeletin-go")

check_header "Checking dependency licenses (source-based)"

# Check if go-licenses is installed
if ! command -v go-licenses &> /dev/null; then
    check_failure "go-licenses not installed" "" \
        "Install with: go install github.com/google/go-licenses/v2@latest"$'\n'"Or run: task setup"
    exit 1
fi

# Run license check
if run_check "go-licenses check --allowed_licenses='$ALLOWED_LICENSES' --ignore='$MODULE_PATH' ./... 2>&1"; then
    check_success "All dependency licenses compliant"
    check_note "Source-based check. Run 'task check:license:binary' for release verification."
    exit 0
else
    check_failure \
        "License compliance check failed (source-based)" \
        "$CHECK_OUTPUT" \
        "Remove dependency: go get <package>@none"$'\n'"Find alternative: Search pkg.go.dev for MIT/Apache-2.0 alternatives"$'\n'"Review policy: See docs/licenses.md for customization"$'\n'"Generate report: task generate:license:report"
    exit 1
fi
