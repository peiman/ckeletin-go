#!/bin/bash
# Check if config constants are up-to-date with the registry
# This script ensures that internal/config/keys_generated.go is in sync with the config registry

set -e

echo "🔍 Checking if config constants are up-to-date..."

# Generate to temp file
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

# Run full generation task (includes formatting)
# Suppress output and capture the file before git restore
go run scripts/generate-config-constants.go > /dev/null 2>&1
./scripts/format-go.sh fix internal/config/keys_generated.go > /dev/null 2>&1

# Copy newly generated and formatted file to temp
cp internal/config/keys_generated.go "$TEMP_FILE"

# Restore original from git
git checkout internal/config/keys_generated.go 2>/dev/null || true

# Compare
if ! diff -q "$TEMP_FILE" internal/config/keys_generated.go > /dev/null 2>&1; then
    echo "❌ ERROR: Config constants are out of date!"
    echo ""
    echo "The generated constants in internal/config/keys_generated.go do not match the current registry."
    echo ""
    echo "Please run:"
    echo "  task generate:constants"
    echo ""
    echo "Then commit the updated file."
    exit 1
fi

echo "✅ Config constants are up-to-date"
