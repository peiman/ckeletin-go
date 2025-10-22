#!/bin/bash
# validate-command-patterns.sh
#
# Validates that command files follow ckeletin-go ultra-thin command patterns.
# This script checks for common violations and can be run in CI to enforce consistency.
#
# Whitelist mechanism: Add // ckeletin:allow-custom-command to bypass checks

set -e

ERRORS=0
WARNINGS=0

# Colors for output
RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "🔍 Validating command patterns..."
echo ""

# Get all command files (exclude framework files and tests)
COMMAND_FILES=$(find cmd -name "*.go" -not -name "*_test.go" -not -name "root.go" -not -name "flags.go" -not -name "helpers.go" -not -name "template*.go")

for cmd_file in $COMMAND_FILES; do
    cmd_name=$(basename "$cmd_file" .go)

    echo "Checking $cmd_name..."

    # Check if file has whitelist comment
    if grep -q "// ckeletin:allow-custom-command" "$cmd_file"; then
        echo "  ℹ️  Whitelisted (custom command pattern allowed)"
        continue
    fi

    # Check 1: Command metadata exists
    if ! find internal/config/commands -name "${cmd_name}_config.go" 2>/dev/null | grep -q .; then
        echo -e "  ${RED}❌ Missing${NC} metadata file: internal/config/commands/${cmd_name}_config.go"
        ((ERRORS++))
    fi

    # Check 2: Uses NewCommand helper
    if ! grep -q "NewCommand(" "$cmd_file"; then
        echo -e "  ${YELLOW}⚠️  Warning${NC}: Does not use NewCommand() helper"
        echo "     Consider using: var ${cmd_name}Cmd = NewCommand(commands.MetadataName, run${cmd_name})"
        ((WARNINGS++))
    fi

    # Check 3: Uses MustAddToRoot helper
    if ! grep -q "MustAddToRoot(" "$cmd_file"; then
        if grep -q "RootCmd.AddCommand" "$cmd_file" && grep -q "setupCommandConfig" "$cmd_file"; then
            echo -e "  ${YELLOW}⚠️  Warning${NC}: Manual RootCmd setup"
            echo "     Consider using: MustAddToRoot(${cmd_name}Cmd)"
            ((WARNINGS++))
        fi
    fi

    # Check 4: Business logic detection (simple heuristic)
    # Look for complex control flow outside of run* functions
    if grep -v "^func run" "$cmd_file" | grep -E "(for\s+.*{|if\s+.*{\s*$|switch\s+.*{)" | grep -v "^//" | grep -q .; then
        echo -e "  ${YELLOW}⚠️  Warning${NC}: Possible business logic in command file"
        echo "     Business logic should be in internal/${cmd_name}/"
        ((WARNINGS++))
    fi

    # Check 5: File length check (should be ~20-30 lines for ultra-thin)
    line_count=$(wc -l < "$cmd_file")
    if [ "$line_count" -gt 80 ]; then
        echo -e "  ${YELLOW}⚠️  Warning${NC}: Command file is ${line_count} lines (expected ~20-30)"
        echo "     Consider moving logic to internal/${cmd_name}/"
        ((WARNINGS++))
    fi

    if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} Passes all checks"
    fi

    echo ""
done

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ All commands follow the pattern!${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠️  ${WARNINGS} warning(s) found${NC}"
    echo "Warnings are suggestions and won't fail the build."
    exit 0
else
    echo -e "${RED}❌ ${ERRORS} error(s) found, ${WARNINGS} warning(s)${NC}"
    echo ""
    echo "To whitelist a command from validation, add this comment to the file:"
    echo "  // ckeletin:allow-custom-command"
    exit 1
fi
