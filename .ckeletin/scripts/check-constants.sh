#!/bin/bash
# Check if config constants are up-to-date with the registry
# This script ensures that internal/config/keys_generated.go is in sync with the config registry

set -eo pipefail

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Validating ADR-005: Config constants in sync"

# Temp files for comparison (.go names so the formatter treats them as Go)
TEMP_DIR=$(mktemp -d)
TEMP_CURRENT="${TEMP_DIR}/keys_current.go"
TEMP_FRESH="${TEMP_DIR}/keys_generated.go"
KEYS_FILE=""

# The generator honors CKELETIN_CONSTANTS_OUT, so the working tree is never
# mutated. Legacy project-owned generators (old structure) may ignore the
# override and overwrite the real file; the EXIT trap restores it even if
# the script is interrupted mid-run.
restore_and_cleanup() {
    if [ -n "$KEYS_FILE" ] && [ -s "$TEMP_CURRENT" ] && ! cmp -s "$TEMP_CURRENT" "$KEYS_FILE"; then
        cp "$TEMP_CURRENT" "$KEYS_FILE"
    fi
    rm -rf "$TEMP_DIR"
}
trap restore_and_cleanup EXIT

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

# Generate fresh constants to the temp file
if ! GEN_OUTPUT=$(CKELETIN_CONSTANTS_OUT="$TEMP_FRESH" go run "$GEN_SCRIPT" 2>&1); then
    check_failure \
        "Failed to generate fresh config constants" \
        "$GEN_OUTPUT" \
        "Fix the generator error above, then re-run: task check"
    exit 1
fi

# Legacy generator without CKELETIN_CONSTANTS_OUT support wrote to the real
# file instead; copy its output and let the EXIT trap restore the original
if [ ! -s "$TEMP_FRESH" ]; then
    cp "$KEYS_FILE" "$TEMP_FRESH"
fi

if ! FMT_OUTPUT=$(task ckeletin:format:staged -- "$TEMP_FRESH" 2>&1); then
    check_failure \
        "Failed to format freshly generated constants" \
        "$FMT_OUTPUT" \
        "Fix the formatting error above, then re-run: task check"
    exit 1
fi

# Compare: current working tree should match freshly generated
if ! diff -q "$TEMP_CURRENT" "$TEMP_FRESH" > /dev/null 2>&1; then
    check_failure \
        "Config constants are out of date" \
        "Generated constants in keys_generated.go don't match the registry" \
        "Run: task ckeletin:generate:config:key-constants"$'\n'"Then commit the updated file"
    exit 1
fi

check_success "Config constants are up-to-date"
