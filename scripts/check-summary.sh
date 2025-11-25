#!/usr/bin/env bash
# Display summary after all checks pass
# This script only runs if all checks succeeded (task stops on first failure)

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

echo ""
echo "$SEPARATOR"
if [ "$CHECK_MODE" = "fast" ]; then
  echo "✅ All fast checks passed"
else
  echo "✅ All checks passed (16/16)"
fi
echo "$SEPARATOR"
echo ""
echo "✅ Development tools installed"
echo "✅ Code formatting"
echo "✅ Linting"
echo "✅ ADR-001: Ultra-thin command pattern"
echo "✅ ADR-002: Config defaults in registry"
echo "✅ ADR-002: Type-safe config consumption"
echo "✅ ADR-005: Config constants in sync"
echo "✅ ADR-008: Architecture SSOT"
echo "✅ ADR-009: Layered architecture"
echo "✅ ADR-010: Package organization"
echo "✅ ADR-012: Output patterns"

if [ "$CHECK_MODE" != "fast" ]; then
  echo "✅ Dependency integrity"
  echo "✅ No security vulnerabilities"
  echo "✅ License compliance (source)"
  echo "✅ License compliance (binary)"
  echo "✅ All tests passing"
else
  echo "✅ Tests passing (unit only)"
fi

echo ""
echo "$SEPARATOR"
echo "🚀 Ready to commit!"
echo "$SEPARATOR"
echo ""
