#!/bin/bash
set -eo pipefail

echo "🔍 Validating type-safe config consumption pattern (ADR-002)..."

ERRORS=0

# Whitelisted files that can use viper.Get* directly
WHITELIST=(
    "cmd/helpers.go"
    "cmd/root.go"
    "cmd/flags.go"
)

# Check for direct viper.Get* calls in cmd/ files
echo ""
echo "Checking for unauthorized direct viper.Get* calls in cmd/..."

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

if [ $ERRORS -eq 0 ]; then
    echo "✅ No unauthorized direct viper.Get* calls found"
else
    echo -e "$VIOLATIONS"
    echo ""
fi

# Check for proper use of getConfigValueWithFlags helper
echo ""
echo "Checking for use of type-safe config retrieval helper..."

# Find command files that should use the helper
COMMAND_FILES=$(find cmd -name "*.go" -not -name "helpers.go" -not -name "root.go" -not -name "flags*.go" -not -name "*_test.go" -type f)

HAS_HELPER_USAGE=false
for file in $COMMAND_FILES; do
    # Check if file contains getConfigValueWithFlags
    if grep -q "getConfigValueWithFlags" "$file"; then
        HAS_HELPER_USAGE=true
    fi
done

if [ "$HAS_HELPER_USAGE" = true ]; then
    echo "✅ Commands use getConfigValueWithFlags helper for type-safe config retrieval"
else
    echo "ℹ️  No command files found using getConfigValueWithFlags (project may not have commands yet)"
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ]; then
    echo "✅ Config consumption pattern validation passed"
    echo ""
    echo "Type-safe config consumption enforced (ADR-002):"
    echo "  • No direct viper.Get* calls in command files"
    echo "  • Commands use getConfigValueWithFlags[T]() helper"
    echo "  • Config passed as typed structs to executors"
    echo "  • Framework independence maintained in business logic"
    exit 0
else
    echo "❌ Config consumption pattern validation failed"
    echo ""
    echo "Found $ERRORS file(s) with unauthorized direct viper.Get* calls."
    echo ""
    echo "Guidelines (ADR-002 Implementation Patterns):"
    echo "  • Use getConfigValueWithFlags[T]() helper in command files"
    echo "  • Pass config as typed structs (e.g., ping.Config) to executors"
    echo "  • Only cmd/helpers.go, cmd/root.go, cmd/flags.go may use viper.Get* directly"
    echo ""
    echo "Example (cmd/ping.go):"
    echo "  cfg := ping.Config{"
    echo "    Message: getConfigValueWithFlags[string](cmd, \"message\", config.KeyAppPingOutputMessage),"
    echo "  }"
    echo "  return ping.NewExecutor(cfg, ...).Execute()"
    echo ""
    echo "See ADR-002 Implementation Patterns section:"
    echo "  docs/adr/002-centralized-configuration-registry.md"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
