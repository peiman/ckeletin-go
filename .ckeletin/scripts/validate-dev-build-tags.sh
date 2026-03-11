#!/usr/bin/env bash
# Validate that cmd/dev*.go files (non-test) have //go:build dev tag
# Enforces ADR-012: Dev Commands Build Tags

set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "ADR-012: Dev build tags"

violations=0

# Find cmd/dev*.go files that are not test files
while IFS= read -r file; do
    [ -z "$file" ] && continue

    # Check if file has //go:build dev tag (must be first non-empty, non-comment line)
    if ! head -5 "$file" | grep -q '//go:build dev'; then
        echo "  ❌ Missing '//go:build dev' tag: $file"
        violations=$((violations + 1))
    fi
done < <(find cmd/ -name 'dev*.go' ! -name '*_test.go' 2>/dev/null)

if [ "$violations" -gt 0 ]; then
    check_failure "dev-build-tags" "$violations file(s) missing //go:build dev tag" \
        "Add '//go:build dev' as the first line of each cmd/dev*.go file"
    exit 1
fi

check_success "dev build tags validated"
