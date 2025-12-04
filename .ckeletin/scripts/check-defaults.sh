#!/bin/bash
# scripts/check-defaults.sh
#
# This script checks for unauthorized direct calls to viper.SetDefault()
# outside of the internal/config/registry.go file.
# Test files (*_test.go) are exempted from this rule.

# Set strict mode
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Validating ADR-002: Config defaults in registry"

# Find all Go files that call viper.SetDefault(), excluding:
# 1. registry.go (authorized location)
# 2. *_test.go files (allowed in tests)
# 3. comment lines (not actual calls)
UNAUTHORIZED_DEFAULTS=$(grep -rn --include="*.go" --exclude="*_test.go" "viper\.SetDefault" . | grep -v "internal/config/registry.go" | grep -v "//.*viper\.SetDefault" || true)

if [ -n "$UNAUTHORIZED_DEFAULTS" ]; then
    check_failure \
        "Found unauthorized viper.SetDefault() calls" \
        "$UNAUTHORIZED_DEFAULTS" \
        "All defaults must be defined in internal/config/registry.go"$'\n'"Move these defaults to the registry"
    exit 1
else
    check_success "No unauthorized viper.SetDefault() calls"
fi 