#!/bin/bash
# Check if system Go version matches .go-version (SSOT)
# Also validates .go-version meets go.mod minimum requirement
#
# Usage:
#   ./scripts/check-go-version.sh         # Strict match (for CI)
#   ./scripts/check-go-version.sh --minor # Minor version match (for dev)

set -e

GO_VERSION_FILE=".go-version"
GO_MOD_FILE="go.mod"

# Parse versions into components (major minor patch)
parse_version() {
    local version="$1"
    # Handle both "1.25" and "1.25.5" formats
    echo "$version" | sed -E 's/^([0-9]+)\.([0-9]+)(\.([0-9]+))?$/\1 \2 \4/'
}

# Compare versions: returns 0 if v1 >= v2, 1 otherwise
version_gte() {
    local v1_major=$1 v1_minor=$2 v1_patch=${3:-0}
    local v2_major=$4 v2_minor=$5 v2_patch=${6:-0}

    if [ "$v1_major" -gt "$v2_major" ]; then return 0; fi
    if [ "$v1_major" -lt "$v2_major" ]; then return 1; fi
    if [ "$v1_minor" -gt "$v2_minor" ]; then return 0; fi
    if [ "$v1_minor" -lt "$v2_minor" ]; then return 1; fi
    if [ "$v1_patch" -ge "$v2_patch" ]; then return 0; fi
    return 1
}

# Check .go-version file exists
if [ ! -f "$GO_VERSION_FILE" ]; then
    echo "❌ $GO_VERSION_FILE not found"
    echo "   This file is the SSOT for Go version"
    exit 1
fi

# Read expected version from .go-version (trim whitespace)
EXPECTED_VERSION=$(tr -d '[:space:]' < "$GO_VERSION_FILE")

if [ -z "$EXPECTED_VERSION" ]; then
    echo "❌ $GO_VERSION_FILE is empty"
    exit 1
fi

# Get minimum version from go.mod
if [ -f "$GO_MOD_FILE" ]; then
    MIN_VERSION=$(grep -E '^go [0-9]+\.[0-9]+' "$GO_MOD_FILE" | awk '{print $2}')
fi

# Get current Go version (e.g., "go1.25.5" -> "1.25.5")
CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/^go//')

if [ -z "$CURRENT_VERSION" ]; then
    echo "❌ Could not determine Go version"
    echo "   Is Go installed?"
    exit 1
fi

# Parse all versions
read -r EXPECTED_MAJOR EXPECTED_MINOR EXPECTED_PATCH <<< "$(parse_version "$EXPECTED_VERSION")"
read -r CURRENT_MAJOR CURRENT_MINOR CURRENT_PATCH <<< "$(parse_version "$CURRENT_VERSION")"

# Default patch to 0 if not specified
EXPECTED_PATCH="${EXPECTED_PATCH:-0}"
CURRENT_PATCH="${CURRENT_PATCH:-0}"

# Check .go-version meets go.mod minimum
if [ -n "$MIN_VERSION" ]; then
    read -r MIN_MAJOR MIN_MINOR MIN_PATCH <<< "$(parse_version "$MIN_VERSION")"
    MIN_PATCH="${MIN_PATCH:-0}"

    if ! version_gte "$EXPECTED_MAJOR" "$EXPECTED_MINOR" "$EXPECTED_PATCH" "$MIN_MAJOR" "$MIN_MINOR" "$MIN_PATCH"; then
        echo "❌ .go-version ($EXPECTED_VERSION) is below go.mod minimum ($MIN_VERSION)"
        echo "   Update .go-version to at least $MIN_VERSION"
        exit 1
    fi
fi

# Check mode
MODE="${1:-strict}"

case "$MODE" in
    --minor)
        # Minor version match: 1.25.x matches 1.25.anything
        if [ "$CURRENT_MAJOR" = "$EXPECTED_MAJOR" ] && [ "$CURRENT_MINOR" = "$EXPECTED_MINOR" ]; then
            echo "✅ Go version OK: $CURRENT_VERSION (expected ~$EXPECTED_MAJOR.$EXPECTED_MINOR, min: ${MIN_VERSION:-n/a})"
            exit 0
        else
            echo "❌ Go version mismatch"
            echo "   Expected: ~$EXPECTED_MAJOR.$EXPECTED_MINOR.x (from .go-version)"
            echo "   Minimum:  $MIN_VERSION (from go.mod)"
            echo "   Found:    $CURRENT_VERSION"
            echo ""
            echo "💡 Update your Go installation or update .go-version"
            exit 1
        fi
        ;;
    --strict|*)
        # Strict match: exact version required
        if [ "$CURRENT_VERSION" = "$EXPECTED_VERSION" ]; then
            echo "✅ Go version OK: $CURRENT_VERSION (min: ${MIN_VERSION:-n/a})"
            exit 0
        else
            echo "❌ Go version mismatch"
            echo "   Expected: $EXPECTED_VERSION (from .go-version)"
            echo "   Minimum:  ${MIN_VERSION:-n/a} (from go.mod)"
            echo "   Found:    $CURRENT_VERSION"
            echo ""
            echo "💡 Options:"
            echo "   1. Update Go to $EXPECTED_VERSION"
            echo "   2. Update .go-version to $CURRENT_VERSION (if intentional)"
            echo ""
            echo "   After upgrading Go, rebuild tools: task setup"
            exit 1
        fi
        ;;
esac
