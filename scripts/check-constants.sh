#!/bin/bash
# Check if config constants are up-to-date with the registry
# This script ensures that internal/config/keys_generated.go is in sync with the config registry

set -e

# Source standard output functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Validating ADR-005: Config constants in sync"

# Generate to temp file
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

# Run full generation task (includes formatting)
# Suppress output and capture the file before git restore
go run scripts/generate-config-constants.go > /dev/null 2>&1
task format:staged -- internal/config/keys_generated.go > /dev/null 2>&1

# Copy newly generated and formatted file to temp
cp internal/config/keys_generated.go "$TEMP_FILE"

# Restore original from git
git checkout internal/config/keys_generated.go 2>/dev/null || true

# Compare
if ! diff -q "$TEMP_FILE" internal/config/keys_generated.go > /dev/null 2>&1; then
    check_failure \
        "Config constants are out of date" \
        "Generated constants in keys_generated.go don't match the registry" \
        "Run: task generate:config:key-constants"$'\n'"Then commit the updated file"
    exit 1
fi

check_success "Config constants are up-to-date"
