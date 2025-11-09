#!/bin/bash
set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Validating package organization (ADR-010)"

ERRORS=0
ERROR_DETAILS=""

# Rule 1: No pkg/ directory with Go files
if [ -d "pkg" ]; then
    GO_FILES_IN_PKG=$(find pkg -name "*.go" -not -name "*_test.go" 2>/dev/null || true)
    if [ -n "$GO_FILES_IN_PKG" ]; then
        ERROR_DETAILS+="Found Go files in pkg/ directory:"$'\n'
        ERROR_DETAILS+="$GO_FILES_IN_PKG"$'\n\n'
        ERROR_DETAILS+="ckeletin-go is a CLI application, not a library."$'\n'
        ERROR_DETAILS+="All implementation should be in internal/ to prevent external imports."$'\n'
        ERRORS=$((ERRORS + 1))
    fi
fi

# Rule 2: Only main.go and main_test.go at root
UNAUTHORIZED_ROOT_FILES=$(find . -maxdepth 1 -name "*.go" ! -name "main.go" ! -name "main_test.go" 2>/dev/null || true)
if [ -n "$UNAUTHORIZED_ROOT_FILES" ]; then
    ERROR_DETAILS+="Found unauthorized .go files at project root:"$'\n'
    ERROR_DETAILS+="$UNAUTHORIZED_ROOT_FILES"$'\n\n'
    ERROR_DETAILS+="Only main.go and main_test.go are allowed at root."$'\n'
    ERROR_DETAILS+="Move other files to cmd/ or internal/."$'\n'
    ERRORS=$((ERRORS + 1))
fi

# Rule 3: All Go packages in expected locations
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
    ERROR_DETAILS+="Found Go packages in unauthorized locations:"$'\n'
    ERROR_DETAILS+="$INVALID_DIRS"$'\n\n'
    ERROR_DETAILS+="Go packages must be in: cmd/, internal/, scripts/, test/, or testdata/"$'\n'
    ERRORS=$((ERRORS + 1))
fi

# Rule 4: main.go exists (sanity check)
if [ ! -f "main.go" ]; then
    ERROR_DETAILS+="main.go not found at project root"$'\n\n'
    ERROR_DETAILS+="CLI applications must have an entry point at main.go"$'\n'
    ERRORS=$((ERRORS + 1))
fi

# Summary
if [ $ERRORS -eq 0 ]; then
    echo "$SEPARATOR"
    echo "✅ Package organization validation passed"
    echo ""
    echo "Structure follows ADR-010 (CLI-first organization):"
    echo "  • No public API surface (no pkg/)"
    echo "  • All implementation in internal/"
    echo "  • CLI interface via cmd/"
    echo "  • Clean project root"
    echo "$SEPARATOR"
    exit 0
else
    REMEDIATION="Fix issues to maintain CLI-first architecture"$'\n'
    REMEDIATION+="ckeletin-go is a CLI application, not a library"$'\n'
    REMEDIATION+="All implementation goes in internal/ (private)"$'\n'
    REMEDIATION+="Commands go in cmd/ (CLI interface)"$'\n'
    REMEDIATION+="Keep project root clean (only main.go allowed)"$'\n'
    REMEDIATION+="See ADR-010: docs/adr/010-package-organization-strategy.md"

    check_failure \
        "Package organization validation failed" \
        "$ERROR_DETAILS" \
        "$REMEDIATION"
    exit 1
fi
