#!/usr/bin/env bash
# Generate NOTICE file with dependency attributions
# Required for Apache-2.0 dependencies and binary distributions
# See: ADR-011 License Compliance Strategy

set -e

OUTPUT="NOTICE"
CSV_REPORT="reports/licenses.csv"

echo "📄 Generating attribution NOTICE file..."
echo "   Source: $CSV_REPORT"
echo "   Output: $OUTPUT"
echo ""

# Ensure report exists
if [ ! -f "$CSV_REPORT" ]; then
    echo "⚠️  License report not found, generating..."
    echo ""
    if ! ./scripts/generate-license-report.sh; then
        echo "❌ Failed to generate license report"
        exit 1
    fi
    echo ""
fi

# Check if report is empty (only header)
LINE_COUNT=$(wc -l < "$CSV_REPORT")
if [ "$LINE_COUNT" -le 1 ]; then
    echo "⚠️  No dependencies found in report"
    echo "   Skipping NOTICE generation"
    exit 0
fi

# Get project info
PROJECT_NAME=$(go list -m 2>/dev/null | sed 's|.*/||' || echo "ckeletin-go")
YEAR=$(date +%Y)

# Generate NOTICE header
cat > "$OUTPUT" <<EOF
NOTICE

This software includes the following third-party components.

Complete license texts are available in the third_party/licenses/ directory
or can be retrieved via: task generate:license:files

================================================================================

EOF

# Parse CSV and generate entries (skip header line)
ENTRY_COUNT=0
tail -n +2 "$CSV_REPORT" | while IFS=',' read -r package url license; do
    # Clean up fields (remove quotes if present)
    package=$(echo "$package" | tr -d '"')
    url=$(echo "$url" | tr -d '"')
    license=$(echo "$license" | tr -d '"')

    # Skip empty lines
    [ -z "$package" ] && continue

    cat >> "$OUTPUT" <<EOF
$package
License: $license
URL: $url

================================================================================

EOF
    ENTRY_COUNT=$((ENTRY_COUNT + 1))
done

# Add footer
cat >> "$OUTPUT" <<EOF

Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')
Total dependencies: $ENTRY_COUNT

For questions about these licenses, see: docs/licenses.md
EOF

# Count final entries
FINAL_COUNT=$(tail -n +2 "$CSV_REPORT" | wc -l)

echo "✅ NOTICE file generated successfully"
echo "   File: $OUTPUT"
echo "   Lines: $(wc -l < $OUTPUT)"
echo "   Dependencies: $FINAL_COUNT"
echo ""
echo "Usage:"
echo "  - Include NOTICE in binary distributions"
echo "  - Required for Apache-2.0 licensed dependencies"
echo "  - Provides attribution for all open-source components"
echo ""
echo "Next steps:"
echo "  - Review: cat $OUTPUT"
echo "  - Include in releases: See .goreleaser.yml"
echo "  - Save license files: task generate:license:files"
