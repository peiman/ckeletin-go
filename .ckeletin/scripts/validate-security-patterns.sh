#!/bin/bash
set -e

# scripts/validate-security-patterns.sh
# Enforces ADR-004: Security Validation in Configuration
#
# Validates that security patterns are properly implemented:
# 1. Security constants are defined in limits.go
# 2. Security validation functions exist in security.go
# 3. Security validation is called during initialization
# 4. Integration tests cover security scenarios

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo "🔍 Validating ADR-004: Security validation patterns..."

ERRORS=0

# Determine config paths (framework vs old structure)
if [ -f ".ckeletin/pkg/config/limits.go" ]; then
    CONFIG_DIR=".ckeletin/pkg/config"
else
    CONFIG_DIR="internal/config"
fi

# 1. Check security constants are defined
if ! grep -q "MaxConfigFileSize" "$CONFIG_DIR/limits.go" 2>/dev/null; then
    echo -e "${RED}❌ Missing MaxConfigFileSize constant in $CONFIG_DIR/limits.go${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "  ${GREEN}✓${NC} MaxConfigFileSize constant defined"
fi

if ! grep -q "MaxStringValueLength" "$CONFIG_DIR/limits.go" 2>/dev/null; then
    echo -e "${RED}❌ Missing MaxStringValueLength constant in $CONFIG_DIR/limits.go${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "  ${GREEN}✓${NC} MaxStringValueLength constant defined"
fi

if ! grep -q "MaxSliceLength" "$CONFIG_DIR/limits.go" 2>/dev/null; then
    echo -e "${RED}❌ Missing MaxSliceLength constant in $CONFIG_DIR/limits.go${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "  ${GREEN}✓${NC} MaxSliceLength constant defined"
fi

# 2. Check security validation functions exist
if ! grep -q "func ValidateConfigFileSecurity" "$CONFIG_DIR/security.go" 2>/dev/null; then
    echo -e "${RED}❌ Security validation function missing in $CONFIG_DIR/security.go${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "  ${GREEN}✓${NC} ValidateConfigFileSecurity function exists"
fi

if ! grep -q "func ValidateConfigFilePermissions" "$CONFIG_DIR/security.go" 2>/dev/null; then
    echo -e "${RED}❌ ValidateConfigFilePermissions function missing in $CONFIG_DIR/security.go${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "  ${GREEN}✓${NC} ValidateConfigFilePermissions function exists"
fi

if ! grep -q "func ValidateConfigFileSize" "$CONFIG_DIR/security.go" 2>/dev/null; then
    echo -e "${RED}❌ ValidateConfigFileSize function missing in $CONFIG_DIR/security.go${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "  ${GREEN}✓${NC} ValidateConfigFileSize function exists"
fi

# 3. Check security validation is called in root.go
if ! grep -q "ValidateConfigFileSecurity\|ValidateConfigFilePermissions" cmd/root.go 2>/dev/null; then
    echo -e "${RED}❌ Security validation not called in cmd/root.go${NC}"
    echo "   Config loading should call config.ValidateConfigFileSecurity() before loading"
    ERRORS=$((ERRORS + 1))
else
    echo -e "  ${GREEN}✓${NC} Security validation called during config loading"
fi

# 4. Check integration tests exist for security scenarios
if [ -f "test/integration/error_scenarios_test.go" ]; then
    echo -e "  ${GREEN}✓${NC} Security error scenario tests exist"
else
    echo -e "${YELLOW}⚠️  Warning: test/integration/error_scenarios_test.go not found${NC}"
fi

# 5. Check security test file exists
if [ -f "$CONFIG_DIR/security_test.go" ]; then
    echo -e "  ${GREEN}✓${NC} Security unit tests exist"
else
    echo -e "${YELLOW}⚠️  Warning: $CONFIG_DIR/security_test.go not found${NC}"
fi

# 6. Check limits test file exists
if [ -f "$CONFIG_DIR/limits_test.go" ]; then
    echo -e "  ${GREEN}✓${NC} Limits unit tests exist"
else
    echo -e "${YELLOW}⚠️  Warning: $CONFIG_DIR/limits_test.go not found${NC}"
fi

echo ""

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}❌ Security validation check failed with $ERRORS error(s)${NC}"
    echo ""
    echo "ADR-004 requires:"
    echo "  1. Security constants in $CONFIG_DIR/limits.go"
    echo "  2. Validation functions in $CONFIG_DIR/security.go"
    echo "  3. Security validation called during config load in cmd/root.go"
    exit 1
fi

echo -e "${GREEN}✅ All security patterns compliant with ADR-004${NC}"
