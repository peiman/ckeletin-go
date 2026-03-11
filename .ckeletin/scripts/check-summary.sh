#!/usr/bin/env bash
# Display summary after all checks pass
# This script only runs if all checks succeeded (task stops on first failure)

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

CHECKS=()

CHECKS+=("Code formatting")
CHECKS+=("Linting")
CHECKS+=("ADR-001: Ultra-thin command pattern")
CHECKS+=("ADR-002: Config defaults in registry")
CHECKS+=("ADR-002: Type-safe config consumption")
CHECKS+=("ADR-004: Security patterns")
CHECKS+=("ADR-005: Config constants in sync")
CHECKS+=("ADR-000: Task naming")
CHECKS+=("ADR-008: Architecture SSOT")
CHECKS+=("ADR-009: Layered architecture")
CHECKS+=("ADR-010: Package organization")
CHECKS+=("ADR-013: Output patterns")
CHECKS+=("ADR-012: Dev build tags")

if [ "$CHECK_MODE" != "fast" ]; then
  CHECKS+=("No hardcoded secrets (gitleaks)")
  CHECKS+=("Static analysis passed (semgrep)")
  CHECKS+=("Dependency integrity")
  CHECKS+=("License compliance (source)")
  CHECKS+=("License compliance (binary)")
  CHECKS+=("SBOM vulnerability scan")
  CHECKS+=("All tests passing (unit + integration + race)")
else
  CHECKS+=("Tests passing (unit only)")
fi

COUNT=${#CHECKS[@]}

echo ""
echo "$SEPARATOR"
if [ "$CHECK_MODE" = "fast" ]; then
  echo "✅ All fast checks passed (${COUNT}/${COUNT})"
else
  echo "✅ All checks passed (${COUNT}/${COUNT})"
fi
echo "$SEPARATOR"
echo ""

for check in "${CHECKS[@]}"; do
  echo "✅ $check"
done

echo ""
echo "$SEPARATOR"
echo "🚀 Ready to commit!"
echo "$SEPARATOR"
echo ""
