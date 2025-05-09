#!/bin/bash
# scripts/check-defaults.sh
#
# This script checks for unauthorized direct calls to viper.SetDefault()
# outside of the internal/config/registry.go file.
# Test files (*_test.go) are exempted from this rule.

# Set strict mode
set -eo pipefail

echo "Checking for unauthorized viper.SetDefault() calls..."

# Find all Go files that call viper.SetDefault(), excluding:
# 1. registry.go (authorized location)
# 2. *_test.go files (allowed in tests)
# 3. comment lines (not actual calls)
UNAUTHORIZED_DEFAULTS=$(grep -rn --include="*.go" --exclude="*_test.go" "viper\.SetDefault" . | grep -v "internal/config/registry.go" | grep -v "//.*viper\.SetDefault" || true)

if [ -n "$UNAUTHORIZED_DEFAULTS" ]; then
    echo "ERROR: Found unauthorized viper.SetDefault() calls in the following locations:"
    echo ""
    echo "$UNAUTHORIZED_DEFAULTS"
    echo ""
    echo "IMPORTANT: All defaults must be defined ONLY in internal/config/registry.go"
    echo "Please move these defaults to the registry and remove the direct calls."
    exit 1
else
    echo "✅ No unauthorized viper.SetDefault() calls found."
fi 