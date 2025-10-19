#!/usr/bin/env bash

# validate-configuration-patterns.sh
#
# Validates that all command files follow ckeletin-go configuration best practices
#
# This script checks for:
# 1. No direct viper.SetDefault() calls in command files
# 2. All NewXConfig functions accept cmd *cobra.Command parameter
# 3. No manual precedence logic (viper.Get + cmd.Flags().Changed)
# 4. Use of getConfigValue instead of manual checks
# 5. Proper options pattern compliance
#
# Exit codes:
#   0 - All checks passed
#   1 - Violations found

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_FILES=0
VIOLATIONS=0
WARNINGS=0

echo -e "${BLUE}🔍 Validating configuration patterns...${NC}"
echo -e "${BLUE}📋 Checking command files for pattern compliance...${NC}"

# Check 1: No direct viper.SetDefault() in command files
check_viper_set_default() {
    local file=$1
    local matches
    # Filter out comment lines (lines containing // before viper.SetDefault)
    matches=$(grep -n "viper\.SetDefault" "$file" 2>/dev/null | grep -v "//.*viper\.SetDefault" || true)

    if [ -n "$matches" ]; then
        echo -e "  ${RED}❌ $file: Direct viper.SetDefault() calls found${NC}"
        echo "$matches" | while IFS= read -r line; do
            echo -e "     ${RED}   $line${NC}"
        done
        echo -e "     ${RED}   All defaults MUST be defined in internal/config/registry.go${NC}"
        return 1
    fi
    return 0
}

# Check 2: NewXConfig functions must accept cmd *cobra.Command parameter
check_new_config_signature() {
    local file=$1
    local matches
    matches=$(grep -n "^func New.*Config(" "$file" 2>/dev/null || true)

    if [ -n "$matches" ]; then
        # Check if any NewXConfig function is missing cmd parameter
        while IFS= read -r line; do
            local line_num=$(echo "$line" | cut -d: -f1)
            local func_sig=$(echo "$line" | cut -d: -f2-)

            # Check if signature contains "cmd *cobra.Command"
            if ! echo "$func_sig" | grep -qE "cmd \*cobra\.Command"; then
                echo -e "  ${YELLOW}⚠️  $file:$line_num: NewXConfig missing cmd parameter${NC}"
                echo -e "     ${YELLOW}   Found: $func_sig${NC}"
                echo -e "     ${YELLOW}   Should include: cmd *cobra.Command parameter${NC}"
                return 2
            fi
        done <<< "$matches"
    fi
    return 0
}

# Check 3: No manual precedence logic (viper.Get + cmd.Flags().Changed pattern)
check_manual_precedence() {
    local file=$1
    local has_viper_get
    local has_flags_changed

    has_viper_get=$(grep -nE "viper\.Get(String|Int|Bool)" "$file" 2>/dev/null || true)
    has_flags_changed=$(grep -nE "cmd\.Flags\(\)\.Changed" "$file" 2>/dev/null || true)

    # If both patterns exist in the same file, it's likely manual precedence logic
    if [ -n "$has_viper_get" ] && [ -n "$has_flags_changed" ]; then
        # Check if it's NOT in a NewXConfig function (where getConfigValue should be used)
        local in_new_config
        in_new_config=$(grep -n "^func New.*Config(" "$file" 2>/dev/null || true)

        if [ -z "$in_new_config" ]; then
            # This is in runE function - likely manual precedence
            echo -e "  ${RED}❌ $file: Manual precedence logic detected${NC}"
            echo -e "     ${RED}   Found: viper.Get* + cmd.Flags().Changed() pattern${NC}"
            echo -e "     ${RED}   Should use: getConfigValue[T](cmd, flagName, viperKey)${NC}"
            return 1
        fi
    fi
    return 0
}

# Check 4: runE functions should use NewXConfig, not manual viper.Get
check_rune_pattern() {
    local file=$1
    local in_rune=false
    local line_num=0
    local violations_found=false

    while IFS= read -r line; do
        ((line_num++))

        # Check if we're in a runE function
        if echo "$line" | grep -qE "^func run[A-Z].*\(cmd \*cobra\.Command"; then
            in_rune=true
            continue
        fi

        # Check if we exit the function
        if [ "$in_rune" = true ] && echo "$line" | grep -q "^}"; then
            in_rune=false
            continue
        fi

        # If in runE and using viper.Get directly (not through NewXConfig)
        if [ "$in_rune" = true ] && echo "$line" | grep -q "viper\.Get"; then
            if ! echo "$line" | grep -q "NewConfig"; then
                if [ "$violations_found" = false ]; then
                    echo -e "  ${RED}❌ $file: Direct viper.Get in runE function${NC}"
                    violations_found=true
                fi
                echo -e "     ${RED}   Line $line_num: $(echo "$line" | xargs)${NC}"
                echo -e "     ${RED}   Should use: cfg := NewXConfig(cmd) and access cfg fields${NC}"
            fi
        fi
    done < "$file"

    if [ "$violations_found" = true ]; then
        return 1
    fi
    return 0
}

# Main validation loop
for file in cmd/*.go; do
    # Skip test files
    if [[ "$file" == *_test.go ]]; then
        continue
    fi

    # Skip root.go (contains getConfigValue definition)
    if [[ "$file" == "cmd/root.go" ]]; then
        continue
    fi

    # Skip completion.go (generated)
    if [[ "$file" == "cmd/completion.go" ]]; then
        continue
    fi

    TOTAL_FILES=$((TOTAL_FILES + 1))

    passed=true

    # Run checks
    if ! check_viper_set_default "$file"; then
        passed=false
        VIOLATIONS=$((VIOLATIONS + 1))
    fi

    result=0
    check_new_config_signature "$file" || result=$?
    if [ $result -eq 1 ]; then
        passed=false
        VIOLATIONS=$((VIOLATIONS + 1))
    elif [ $result -eq 2 ]; then
        WARNINGS=$((WARNINGS + 1))
    fi

    if ! check_manual_precedence "$file"; then
        passed=false
        VIOLATIONS=$((VIOLATIONS + 1))
    fi

    if ! check_rune_pattern "$file"; then
        passed=false
        VIOLATIONS=$((VIOLATIONS + 1))
    fi

    if [ "$passed" = true ]; then
        echo -e "  ${GREEN}✅ $file${NC}"
    fi
done

echo ""
echo -e "${BLUE}📊 CONFIGURATION PATTERN VALIDATION REPORT:${NC}"
echo "  Files checked: $TOTAL_FILES"
if [ $VIOLATIONS -gt 0 ]; then
    echo -e "  ${RED}Violations: $VIOLATIONS${NC}"
fi
if [ $WARNINGS -gt 0 ]; then
    echo -e "  ${YELLOW}Warnings: $WARNINGS${NC}"
fi

echo ""

if [ $VIOLATIONS -gt 0 ]; then
    echo -e "${RED}❌ VALIDATION FAILED${NC}"
    echo -e "${RED}   $VIOLATIONS critical violations found that must be fixed.${NC}"
    echo ""
    echo -e "${BLUE}🔧 To fix violations:${NC}"
    echo "   - Replace manual viper.Get + cmd.Flags().Changed with getConfigValue"
    echo "   - Ensure all NewXConfig functions accept cmd *cobra.Command parameter"
    echo "   - Use NewXConfig in runE functions instead of direct viper calls"
    echo "   - Define all defaults in internal/config/registry.go"
    echo ""
    echo -e "${BLUE}📚 See docs/CONFIGURATION_PATTERNS.md for examples${NC}"
    exit 1
fi

if [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}⚠️  VALIDATION PASSED WITH WARNINGS${NC}"
    echo -e "${YELLOW}   $WARNINGS warnings found. Consider reviewing these patterns.${NC}"
    exit 0
fi

echo -e "${GREEN}✅ ALL CONFIGURATION PATTERNS VALIDATED SUCCESSFULLY${NC}"
exit 0
