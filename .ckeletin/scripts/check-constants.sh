#!/bin/bash
# Check if config constants are up-to-date with the registry
# This script ensures that internal/config/keys_generated.go is in sync with the config registry

set -e

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Validating ADR-005: Config constants in sync"

# Temp files for comparison
TEMP_CURRENT=$(mktemp)
TEMP_FRESH=$(mktemp)
trap "rm -f $TEMP_CURRENT $TEMP_FRESH" EXIT

# Determine keys file location (framework vs old structure)
if [ -f ".ckeletin/pkg/config/keys_generated.go" ]; then
    KEYS_FILE=".ckeletin/pkg/config/keys_generated.go"
    GEN_SCRIPT=".ckeletin/scripts/generate-config-constants.go"
else
    KEYS_FILE="internal/config/keys_generated.go"
    GEN_SCRIPT="scripts/generate-config-constants.go"
fi

# Save current working tree version (may have uncommitted changes)
cp "$KEYS_FILE" "$TEMP_CURRENT"

# Generate fresh constants to the actual file
go run "$GEN_SCRIPT" > /dev/null 2>&1
task ckeletin:format:staged -- "$KEYS_FILE" > /dev/null 2>&1

# Save freshly generated version
cp "$KEYS_FILE" "$TEMP_FRESH"

# Restore working tree version (don't lose uncommitted work)
cp "$TEMP_CURRENT" "$KEYS_FILE"

# Compare: current working tree should match freshly generated
if ! diff -q "$TEMP_CURRENT" "$TEMP_FRESH" > /dev/null 2>&1; then
    check_failure \
        "Config constants are out of date" \
        "Generated constants in keys_generated.go don't match the registry" \
        "Run: task generate:config:key-constants"$'\n'"Then commit the updated file"
    exit 1
fi

check_success "Config constants are up-to-date"
