#!/usr/bin/env bash
# Generate license report (CSV format) using go-licenses
# Used for audits, documentation, and NOTICE generation
# See: ADR-011 License Compliance Strategy

set -e

OUTPUT_DIR="${LICENSE_REPORT_DIR:-reports}"
OUTPUT_FILE="$OUTPUT_DIR/licenses.csv"
ERROR_LOG="$OUTPUT_DIR/license-errors.log"

# Get module path to ignore self
MODULE_PATH=$(go list -m 2>/dev/null || echo "github.com/peiman/ckeletin-go")

echo "📊 Generating license report..."
echo "   Tool: go-licenses"
echo "   Output: $OUTPUT_FILE"
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

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate report
echo "Scanning dependencies..."
if go-licenses report \
    --ignore="$MODULE_PATH" \
    ./... > "$OUTPUT_FILE" 2> "$ERROR_LOG"; then

    # Count dependencies
    DEP_COUNT=$(($(wc -l < "$OUTPUT_FILE") - 1))  # Subtract header line

    echo ""
    echo "✅ License report generated successfully"
    echo "   File: $OUTPUT_FILE"
    echo "   Dependencies: $DEP_COUNT"

    # Show sample (first 10 lines)
    if [ "$DEP_COUNT" -gt 0 ]; then
        echo ""
        echo "Sample (first 5 dependencies):"
        echo "----------------------------------------"
        head -n 6 "$OUTPUT_FILE" | column -t -s ','
        if [ "$DEP_COUNT" -gt 5 ]; then
            echo "... and $((DEP_COUNT - 5)) more"
        fi
        echo "----------------------------------------"
    fi

    # Check for errors
    if [ -s "$ERROR_LOG" ]; then
        echo ""
        echo "⚠️  Warnings/errors during scan (see $ERROR_LOG):"
        head -n 5 "$ERROR_LOG"
    else
        rm -f "$ERROR_LOG"
    fi

    echo ""
    echo "Use this report for:"
    echo "  - Compliance audits"
    echo "  - NOTICE file generation (task generate:attribution)"
    echo "  - Dependency documentation"

    exit 0
else
    echo ""
    echo "❌ Failed to generate license report"
    echo ""
    if [ -s "$ERROR_LOG" ]; then
        echo "Errors:"
        cat "$ERROR_LOG"
    fi
    exit 1
fi
