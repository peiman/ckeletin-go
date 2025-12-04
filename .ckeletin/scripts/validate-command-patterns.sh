#!/bin/bash
# validate-command-patterns.sh
#
# Validates that command files follow ckeletin-go ultra-thin command patterns.
# This script checks for common violations and can be run in CI to enforce consistency.
#
# Whitelist mechanism: Add // ckeletin:allow-custom-command to bypass checks

set -e

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

ERRORS=0
WARNINGS=0
ERROR_DETAILS=""

check_header "Validating ADR-001: Ultra-thin command pattern"

# Get all command files (exclude framework files and tests)
COMMAND_FILES=$(find cmd -name "*.go" -not -name "*_test.go" -not -name "root.go" -not -name "flags.go" -not -name "helpers.go" -not -name "template*.go")

for cmd_file in $COMMAND_FILES; do
    cmd_name=$(basename "$cmd_file" .go)

    # Check if file has whitelist comment
    if grep -q "// ckeletin:allow-custom-command" "$cmd_file"; then
        continue
    fi

    # Check 1: Command metadata exists (check both project and framework locations)
    if ! find internal/config/commands -name "${cmd_name}_config.go" 2>/dev/null | grep -q .; then
        # Also check framework location
        if ! find .ckeletin/pkg/config/commands -name "${cmd_name}_config.go" 2>/dev/null | grep -q .; then
            ERROR_DETAILS+="$cmd_name: Missing metadata file internal/config/commands/${cmd_name}_config.go"$'\n'
            ((ERRORS++))
        fi
    fi

    # Check 2: Uses NewCommand helper
    if ! grep -q "NewCommand(" "$cmd_file"; then
        ERROR_DETAILS+="$cmd_name: Does not use NewCommand() helper"$'\n'
        ((WARNINGS++))
    fi

    # Check 3: Uses MustAddToRoot helper
    if ! grep -q "MustAddToRoot(" "$cmd_file"; then
        if grep -q "RootCmd.AddCommand" "$cmd_file" && grep -q "setupCommandConfig" "$cmd_file"; then
            ERROR_DETAILS+="$cmd_name: Manual RootCmd setup (consider MustAddToRoot)"$'\n'
            ((WARNINGS++))
        fi
    fi

    # Check 4: Business logic detection (simple heuristic)
    # Look for complex control flow outside of run* functions
    if grep -v "^func run" "$cmd_file" | grep -E "(for\s+.*{|if\s+.*{\s*$|switch\s+.*{)" | grep -v "^//" | grep -q .; then
        ERROR_DETAILS+="$cmd_name: Possible business logic in command file (should be in internal/$cmd_name/)"$'\n'
        ((WARNINGS++))
    fi

    # Check 5: File length check (should be ~20-30 lines for ultra-thin)
    line_count=$(wc -l < "$cmd_file")
    if [ "$line_count" -gt 80 ]; then
        ERROR_DETAILS+="$cmd_name: Command file is $line_count lines (expected ~20-30)"$'\n'
        ((WARNINGS++))
    fi
done

# Summary
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    check_success "All commands follow the pattern"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    check_success "All commands pass (${WARNINGS} warning(s))"
    check_note "Warnings are suggestions and won't fail the build."
    exit 0
else
    check_failure \
        "${ERRORS} error(s) found, ${WARNINGS} warning(s)" \
        "$ERROR_DETAILS" \
        "To whitelist a command, add: // ckeletin:allow-custom-command"
    exit 1
fi
