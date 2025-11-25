#!/bin/bash
set -e

# scripts/validate-output-patterns.sh
# Enforces ADR-012: Structured Output and Shadow Logging
#
# Rule: Business logic packages (internal/*) must NOT print directly to stdout.
# They must use:
# 1. internal/ui for Data Stream (User output)
# 2. internal/logger for Status/Audit Stream
#
# Exceptions:
# - internal/ui/* (The implementation of the UI layer itself)
# - *_test.go (Tests are allowed to print)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "🔍 Validating ADR-012: Output patterns..."

# Find all Go files in internal/ that are NOT in internal/ui/ and NOT tests
# We search for direct usages of fmt.Print, fmt.Printf, fmt.Println
VIOLATIONS=$(grep -r "fmt\.Print" internal/ \
  --include="*.go" \
  --exclude="*_test.go" \
  --exclude-dir="ui" \
  --exclude-dir="logger" \
  --exclude-dir="testutil" \
  || true)

# Also check for direct os.Stdout usage
STDOUT_VIOLATIONS=$(grep -r "os\.Stdout" internal/ \
  --include="*.go" \
  --exclude="*_test.go" \
  --exclude-dir="ui" \
  --exclude-dir="logger" \
  --exclude-dir="testutil" \
  || true)

if [ -n "$VIOLATIONS" ] || [ -n "$STDOUT_VIOLATIONS" ]; then
    echo -e "${RED}❌ Validation Failed: Direct output detected in business logic.${NC}"
    echo "ADR-012 requires 'internal/' packages to use 'internal/ui' for output or 'log' for status."
    echo ""
    
    if [ -n "$VIOLATIONS" ]; then
        echo "Found direct fmt.Print* calls:"
        echo "$VIOLATIONS"
        echo ""
    fi

    if [ -n "$STDOUT_VIOLATIONS" ]; then
        echo "Found direct os.Stdout usage:"
        echo "$STDOUT_VIOLATIONS"
    fi
    
    exit 1
fi

echo -e "${GREEN}✅ All output patterns compliant with ADR-012${NC}"
