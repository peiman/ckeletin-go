#!/bin/bash
# Validate all example configuration files

set -e

BINARY="./ckeletin-go"
EXAMPLES_DIR="docs/examples"

# Build binary if not exists
if [ ! -f "$BINARY" ]; then
    echo "Building binary..."
    go build -o "$BINARY"
fi

# Validate each example config
EXIT_CODE=0
for config in "$EXAMPLES_DIR"/*.yaml; do
    # Skip README files
    if [[ "$config" == *"README"* ]]; then
        continue
    fi

    echo "Validating $config..."
    if ! $BINARY config validate --file "$config" 2>&1 | tee /tmp/validate_output.txt; then
        echo "❌ Validation failed for $config"
        EXIT_CODE=1
    else
        # Check for warnings (config is valid but has warnings)
        if grep -q "⚠️.*Warnings" /tmp/validate_output.txt; then
            echo "⚠️  Warnings found in $config"
            cat /tmp/validate_output.txt
            EXIT_CODE=1
        else
            echo "✅ $config is valid"
        fi
    fi
    echo
done

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ All example configs are valid"
else
    echo "❌ Some example configs have issues"
fi

exit $EXIT_CODE
