#!/bin/bash
# scripts/validate-layering.sh
#
# Validates that code follows 4-layer architecture (ADR-009)
#
# Enforces:
# - Dependency rules (outer layers depend on inner, never reverse)
# - CLI framework isolation (only cmd/ imports Cobra)
# - Business logic isolation (packages don't import each other)
# - Infrastructure separation (cannot import business logic)
#
# Configuration: .go-arch-lint.yml

set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Validating layered architecture (ADR-009)"

# Check if .go-arch-lint.yml exists
if [ ! -f ".go-arch-lint.yml" ]; then
    echo "❌ Configuration file .go-arch-lint.yml not found"
    echo "   Architecture validation requires configuration file."
    exit 1
fi

# Check if go-arch-lint is installed
if ! command -v go-arch-lint &> /dev/null; then
    echo "📦 go-arch-lint not found, installing..."
    echo ""
    # v1.18.0 is the last release before v1.19.0, which walks with native
    # paths and reports false violations on Windows (fe3dback/go-arch-lint#79).
    if ! go install github.com/fe3dback/go-arch-lint@v1.18.0; then
        echo "❌ Failed to install go-arch-lint"
        echo "   Please install manually: go install github.com/fe3dback/go-arch-lint@v1.18.0"
        exit 1
    fi
    echo "✅ go-arch-lint installed successfully"
    echo ""
fi

# Run the linter (hide output on success, capture on failure)
if run_check "go-arch-lint check 2>&1"; then
    check_success "Layered architecture validation passed"
    echo "  • Entry → Command → Business Logic/Infrastructure"
    echo "  • No reverse dependencies detected"
    echo "  • Cobra isolated to cmd/ layer"
    echo "  • Business logic packages properly isolated"
    exit 0
else
    REMEDIATION="Review layer dependencies and fix violations"$'\n'
    REMEDIATION+=$'\n'"Common issues and solutions:"$'\n'
    REMEDIATION+="  • internal/ package importing from cmd/"$'\n'
    REMEDIATION+="    → Move shared types/interfaces to infrastructure layer"$'\n'
    REMEDIATION+=$'\n'
    REMEDIATION+="  • Business logic importing Cobra"$'\n'
    REMEDIATION+="    → Extract CLI logic to cmd/, keep pure business logic in internal/"$'\n'
    REMEDIATION+=$'\n'
    REMEDIATION+="  • Business logic packages importing each other"$'\n'
    REMEDIATION+="    → Move shared code to infrastructure (e.g., internal/shared/)"$'\n'
    REMEDIATION+="    → Or use interfaces in infrastructure that both packages implement"$'\n'
    REMEDIATION+=$'\n'
    REMEDIATION+="  • Infrastructure importing business logic"$'\n'
    REMEDIATION+="    → Invert the dependency using interfaces"$'\n'
    REMEDIATION+="    → Business logic should import infrastructure, not vice versa"$'\n'
    REMEDIATION+=$'\n'
    REMEDIATION+="See ADR-009: .ckeletin/docs/adr/009-layered-architecture-pattern.md"

    check_failure \
        "Layered architecture validation failed" \
        "$CHECK_OUTPUT" \
        "$REMEDIATION"
    exit 1
fi
