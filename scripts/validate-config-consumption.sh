#!/bin/bash
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Validating ADR-002: Type-safe config consumption"

ERRORS=0

# Whitelisted files that can use viper.Get* directly
WHITELIST=(
    "cmd/helpers.go"
    "cmd/root.go"
    "cmd/flags.go"
)

# Find all .go files in cmd/ excluding whitelisted files and test files
CMD_FILES=$(find cmd -name "*.go" -not -name "*_test.go" -type f)

VIOLATIONS=""
for file in $CMD_FILES; do
    # Check if file is whitelisted
    IS_WHITELISTED=false
    for whitelist_item in "${WHITELIST[@]}"; do
        if [[ "$file" == "$whitelist_item" ]]; then
            IS_WHITELISTED=true
            break
        fi
    done

    if [ "$IS_WHITELISTED" = true ]; then
        continue
    fi

    # Search for viper.Get* calls (GetString, GetBool, GetInt, GetDuration, etc.)
    # Matches: viper.Get, viper.GetString, viper.GetBool, etc.
    VIPER_CALLS=$(grep -n "viper\.Get" "$file" 2>/dev/null || true)

    if [ -n "$VIPER_CALLS" ]; then
        VIOLATIONS="$VIOLATIONS\n\n❌ $file:\n$VIPER_CALLS"
        ERRORS=$((ERRORS + 1))
    fi
done

# Check for proper use of getConfigValueWithFlags helper
COMMAND_FILES=$(find cmd -name "*.go" -not -name "helpers.go" -not -name "root.go" -not -name "flags*.go" -not -name "*_test.go" -type f)

HAS_HELPER_USAGE=false
for file in $COMMAND_FILES; do
    if grep -q "getConfigValueWithFlags" "$file"; then
        HAS_HELPER_USAGE=true
        break
    fi
done

# Summary
if [ $ERRORS -eq 0 ]; then
    check_success "Type-safe config consumption pattern followed"
    exit 0
else
    ERROR_DETAILS="$VIOLATIONS"$'\n\n'"Found $ERRORS file(s) with unauthorized direct viper.Get* calls."

    REMEDIATION="Use getConfigValueWithFlags[T]() helper in command files"$'\n'
    REMEDIATION+="Pass config as typed structs to executors"$'\n'
    REMEDIATION+="Only cmd/helpers.go, cmd/root.go, cmd/flags.go may use viper.Get* directly"$'\n\n'
    REMEDIATION+="Example: cfg := ping.Config{ Message: getConfigValueWithFlags[string](cmd, \"message\", config.Key...) }"$'\n'
    REMEDIATION+="See: docs/adr/002-centralized-configuration-registry.md"

    check_failure \
        "Config consumption pattern validation failed" \
        "$ERROR_DETAILS" \
        "$REMEDIATION"
    exit 1
fi
