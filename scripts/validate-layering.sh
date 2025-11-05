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

echo "🔍 Validating layered architecture (ADR-009)..."
echo ""

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
    if ! go install github.com/fe3dback/go-arch-lint@latest; then
        echo "❌ Failed to install go-arch-lint"
        echo "   Please install manually: go install github.com/fe3dback/go-arch-lint@latest"
        exit 1
    fi
    echo "✅ go-arch-lint installed successfully"
    echo ""
fi

# Run the linter
echo "Running architecture validation..."
echo ""

if go-arch-lint check; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✅ Layered architecture validation passed"
    echo ""
    echo "All layer dependency rules satisfied:"
    echo "  • Entry → Command → Business Logic/Infrastructure"
    echo "  • No reverse dependencies detected"
    echo "  • Cobra isolated to cmd/ layer"
    echo "  • Business logic packages properly isolated"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "❌ Layered architecture validation failed"
    echo ""
    echo "Violations detected in layer dependencies."
    echo ""
    echo "Common issues:"
    echo "  • internal/ package importing from cmd/"
    echo "  • Business logic importing Cobra"
    echo "  • Business logic packages importing each other"
    echo "  • Infrastructure importing business logic"
    echo ""
    echo "See ADR-009 for architecture rules:"
    echo "  docs/adr/009-layered-architecture-pattern.md"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
