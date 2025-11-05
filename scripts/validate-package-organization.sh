#!/bin/bash
set -eo pipefail

echo "🔍 Validating package organization (ADR-010)..."

ERRORS=0

# Rule 1: No pkg/ directory with Go files
echo ""
echo "Checking for Go files in pkg/ directory..."
if [ -d "pkg" ]; then
    GO_FILES_IN_PKG=$(find pkg -name "*.go" -not -name "*_test.go" 2>/dev/null || true)
    if [ -n "$GO_FILES_IN_PKG" ]; then
        echo "❌ Found Go files in pkg/ directory:"
        echo "$GO_FILES_IN_PKG"
        echo ""
        echo "ckeletin-go is a CLI application, not a library."
        echo "All implementation should be in internal/ to prevent external imports."
        echo "See ADR-010: docs/adr/010-package-organization-strategy.md"
        ERRORS=$((ERRORS + 1))
    else
        echo "✅ No Go packages in pkg/ directory"
    fi
else
    echo "✅ No pkg/ directory (correct for CLI-only project)"
fi

# Rule 2: Only main.go and main_test.go at root
echo ""
echo "Checking for unauthorized .go files at root..."
UNAUTHORIZED_ROOT_FILES=$(find . -maxdepth 1 -name "*.go" ! -name "main.go" ! -name "main_test.go" 2>/dev/null || true)
if [ -n "$UNAUTHORIZED_ROOT_FILES" ]; then
    echo "❌ Found unauthorized .go files at project root:"
    echo "$UNAUTHORIZED_ROOT_FILES"
    echo ""
    echo "Only main.go and main_test.go are allowed at root."
    echo "Move other files to cmd/ or internal/."
    echo "See ADR-010: docs/adr/010-package-organization-strategy.md"
    ERRORS=$((ERRORS + 1))
else
    echo "✅ Only main.go and main_test.go at root"
fi

# Rule 3: All Go packages in expected locations
echo ""
echo "Checking that all Go packages are in expected directories..."

# Allowed directories for Go packages
ALLOWED_DIRS=("cmd" "internal" "scripts" "test" "testdata")

# Find all directories containing .go files (excluding vendor and hidden dirs)
ALL_GO_DIRS=$(find . -type f -name "*.go" ! -path "*/vendor/*" ! -path "*/.*/*" ! -path "./main.go" ! -path "./main_test.go" -exec dirname {} \; | sort -u)

# Check each directory
INVALID_DIRS=""
for dir in $ALL_GO_DIRS; do
    # Remove leading ./
    clean_dir=${dir#./}

    # Check if it starts with any allowed directory
    IS_ALLOWED=false
    for allowed in "${ALLOWED_DIRS[@]}"; do
        if [[ "$clean_dir" == "$allowed"* ]]; then
            IS_ALLOWED=true
            break
        fi
    done

    if [ "$IS_ALLOWED" = false ]; then
        INVALID_DIRS="$INVALID_DIRS\n  $clean_dir"
    fi
done

if [ -n "$INVALID_DIRS" ]; then
    echo "❌ Found Go packages in unauthorized locations:"
    echo -e "$INVALID_DIRS"
    echo ""
    echo "Go packages must be in: cmd/, internal/, scripts/, test/, or testdata/"
    echo "See ADR-010: docs/adr/010-package-organization-strategy.md"
    ERRORS=$((ERRORS + 1))
else
    echo "✅ All Go packages in expected directories"
fi

# Rule 4: main.go exists (sanity check)
echo ""
echo "Checking for main.go entry point..."
if [ ! -f "main.go" ]; then
    echo "❌ main.go not found at project root"
    echo ""
    echo "CLI applications must have an entry point at main.go"
    echo "See ADR-010: docs/adr/010-package-organization-strategy.md"
    ERRORS=$((ERRORS + 1))
else
    echo "✅ main.go exists at root"
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ]; then
    echo "✅ Package organization validation passed"
    echo ""
    echo "Structure follows ADR-010 (CLI-first organization):"
    echo "  • No public API surface (no pkg/)"
    echo "  • All implementation in internal/"
    echo "  • CLI interface via cmd/"
    echo "  • Clean project root"
    exit 0
else
    echo "❌ Package organization validation failed"
    echo ""
    echo "Fix the issues above to maintain CLI-first architecture."
    echo ""
    echo "Guidelines:"
    echo "  • ckeletin-go is a CLI application, not a library"
    echo "  • All implementation goes in internal/ (private)"
    echo "  • Commands go in cmd/ (CLI interface)"
    echo "  • Keep project root clean (only main.go allowed)"
    echo ""
    echo "See ADR-010 for rationale:"
    echo "  docs/adr/010-package-organization-strategy.md"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
