#!/usr/bin/env bash
# Check licenses in compiled binary using lichen (binary-based, accurate)
# Analyzes only dependencies actually compiled into the binary
# See: ADR-011 License Compliance Strategy

set -e

# Binary to check (default: ckeletin-go in current directory)
BINARY="${1:-./ckeletin-go}"
CONFIG="${LICENSE_CONFIG:-.lichen.yaml}"

echo "🔍 Checking binary licenses (accurate, release verification)..."
echo "   Tool: lichen"
echo "   Binary: $BINARY"
echo "   Config: $CONFIG"
echo ""

# Check if lichen is installed
if ! command -v lichen &> /dev/null; then
    echo "❌ lichen not installed"
    echo ""
    echo "Install with:"
    echo "  go install github.com/uw-labs/lichen@latest"
    echo ""
    echo "Or run:"
    echo "  task setup"
    exit 1
fi

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo "❌ Binary not found: $BINARY"
    echo ""
    echo "Build the binary first:"
    echo "  task build"
    echo ""
    echo "Or specify binary path:"
    echo "  $0 ./path/to/binary"
    exit 1
fi

# Check if config exists
if [ ! -f "$CONFIG" ]; then
    echo "⚠️  Config not found: $CONFIG"
    echo "   Using lichen defaults (may be permissive)"
    echo ""
fi

# Run lichen
echo "Analyzing binary dependencies..."
if [ -f "$CONFIG" ]; then
    # Use config file
    if lichen --config="$CONFIG" "$BINARY" 2>&1; then
        echo ""
        echo "✅ All binary licenses compliant"
        echo ""
        echo "Note: This check analyzes the compiled binary (accurate for releases)."
        echo "Only dependencies actually shipped are included."
        exit 0
    else
        echo ""
        echo "❌ Binary license compliance check failed"
        echo ""
        echo "Actions:"
        echo "  1. Check which dependency failed (see output above)"
        echo "  2. Remove dependency: go get <package>@none"
        echo "  3. Find alternative: Search for MIT/Apache-2.0 alternatives"
        echo "  4. Add exception: Edit .lichen.yaml exceptions section (if justified)"
        echo ""
        echo "For more info: See docs/licenses.md"
        exit 1
    fi
else
    # No config, use defaults
    if lichen "$BINARY" 2>&1; then
        echo ""
        echo "✅ Binary licenses checked (no config file, lichen defaults used)"
        echo ""
        echo "Recommendation: Create .lichen.yaml for explicit policy enforcement"
        echo "See: docs/licenses.md"
        exit 0
    else
        echo ""
        echo "❌ Binary license check failed"
        echo ""
        echo "Create .lichen.yaml to define your license policy:"
        echo "  See docs/licenses.md for configuration examples"
        exit 1
    fi
fi
