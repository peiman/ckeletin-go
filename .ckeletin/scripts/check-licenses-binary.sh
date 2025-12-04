#!/usr/bin/env bash
# Check licenses in compiled binary using lichen (binary-based, accurate)
# Analyzes only dependencies actually compiled into the binary
# See: ADR-011 License Compliance Strategy

set -e

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

# Binary to check (default: ckeletin-go in current directory)
BINARY="${1:-./ckeletin-go}"
CONFIG="${LICENSE_CONFIG:-.lichen.yaml}"

check_header "Checking binary licenses (release verification)"

# Check if lichen is installed
if ! command -v lichen &> /dev/null; then
    check_failure "lichen not installed" "" \
        "Install with: go install github.com/uw-labs/lichen@latest"$'\n'"Or run: task setup"
    exit 1
fi

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    check_failure "Binary not found: $BINARY" "" \
        "Build the binary first: task build"$'\n'"Or specify binary path: $0 ./path/to/binary"
    exit 1
fi

# Run lichen
if [ -f "$CONFIG" ]; then
    # Use config file
    if run_check "lichen --config='$CONFIG' '$BINARY' 2>&1"; then
        check_success "All binary licenses compliant"
        check_note "Binary analysis includes only shipped dependencies (accurate for releases)."
        exit 0
    else
        check_failure \
            "Binary license compliance check failed" \
            "$CHECK_OUTPUT" \
            "Remove dependency: go get <package>@none"$'\n'"Find alternative: Search for MIT/Apache-2.0 alternatives"$'\n'"Add exception: Edit .lichen.yaml exceptions (if justified)"$'\n'"See: docs/licenses.md"
        exit 1
    fi
else
    # No config, use defaults
    if run_check "lichen '$BINARY' 2>&1"; then
        check_success "Binary licenses checked (lichen defaults)"
        check_note "Recommendation: Create .lichen.yaml for explicit policy. See: docs/licenses.md"
        exit 0
    else
        check_failure \
            "Binary license check failed" \
            "$CHECK_OUTPUT" \
            "Create .lichen.yaml to define your license policy"$'\n'"See: docs/licenses.md for configuration examples"
        exit 1
    fi
fi
