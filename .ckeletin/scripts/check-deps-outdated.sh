#!/usr/bin/env bash
# Check for outdated direct dependencies (informational)
set -eo pipefail

# Check if go-mod-outdated is available
if ! command -v go-mod-outdated &> /dev/null; then
    echo "go-mod-outdated not installed, skipping outdated check"
    exit 0
fi

# Get outdated direct dependencies
OUTDATED=$(go list -u -m -json all 2>/dev/null | go-mod-outdated -update -direct 2>&1 || true)

# Check if there are any outdated deps (skip header line)
OUTDATED_COUNT=$(echo "$OUTDATED" | tail -n +2 | grep -c "^" || true)

if [ "$OUTDATED_COUNT" -gt 0 ]; then
    echo "ℹ $OUTDATED_COUNT direct dependencies have updates available:"
    echo "$OUTDATED" | tail -n +2 | while read -r line; do
        # Parse: MODULE CURRENT LATEST INDIRECT
        MODULE=$(echo "$line" | awk '{print $1}')
        CURRENT=$(echo "$line" | awk '{print $2}')
        LATEST=$(echo "$line" | awk '{print $3}')
        if [ -n "$MODULE" ] && [ -n "$CURRENT" ] && [ -n "$LATEST" ]; then
            echo "  $MODULE $CURRENT → $LATEST"
        fi
    done
fi

exit 0
