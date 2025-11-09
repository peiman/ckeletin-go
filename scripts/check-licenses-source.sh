#!/usr/bin/env bash
# Check dependency licenses using go-licenses (source-based, fast)
# Uses conservative permissive-only policy by default
# See: ADR-011 License Compliance Strategy

set -e

# Default policy: Allow permissive licenses only
# Note: go-licenses doesn't support both --allowed_licenses and --disallowed_types
# We use --allowed_licenses for explicit permissive-only policy
ALLOWED_LICENSES="${LICENSE_ALLOWED:-MIT,Apache-2.0,BSD-2-Clause,BSD-3-Clause,ISC,0BSD,Unlicense}"

# Get module path to ignore self
MODULE_PATH=$(go list -m 2>/dev/null || echo "github.com/peiman/ckeletin-go")

echo "🔍 Checking dependency licenses (source-based, fast)..."
echo "   Tool: go-licenses"
echo "   Allowed: $ALLOWED_LICENSES"
echo ""

# Check if go-licenses is installed
if ! command -v go-licenses &> /dev/null; then
    echo "❌ go-licenses not installed"
    echo ""
    echo "Install with:"
    echo "  go install github.com/google/go-licenses/v2@latest"
    echo ""
    echo "Or run:"
    echo "  task setup"
    exit 1
fi

# Run license check
echo "Scanning dependencies..."
if go-licenses check \
    --allowed_licenses="$ALLOWED_LICENSES" \
    --ignore="$MODULE_PATH" \
    ./... 2>&1; then
    echo ""
    echo "✅ All dependency licenses compliant (source-based check)"
    echo ""
    echo "Note: This is a source-based check (fast, for development)."
    echo "For accurate release verification, run: task check:license:binary"
    exit 0
else
    echo ""
    echo "❌ License compliance check failed (source-based)"
    echo ""
    echo "Actions:"
    echo "  1. Remove dependency: go get <package>@none"
    echo "  2. Find alternative: Search pkg.go.dev for MIT/Apache-2.0 alternatives"
    echo "  3. Override policy: Edit scripts/check-licenses-source.sh if justified"
    echo "  4. Review policy: See docs/licenses.md for customization options"
    echo ""
    echo "For detailed report: task generate:license:report"
    echo "For more info: See docs/licenses.md"
    exit 1
fi
