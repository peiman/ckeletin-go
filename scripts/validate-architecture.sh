#!/bin/bash
# scripts/validate-architecture.sh
#
# Validates that ARCHITECTURE.md maintains SSOT with ADRs
#
# Enforces separation of concerns:
# - ARCHITECTURE.md = Structure (WHAT the system is)
# - ADRs = Decisions (WHY the system is this way)
#
# This prevents documentation drift and duplication

set -e

ARCHITECTURE_FILE="docs/adr/ARCHITECTURE.md"
ADR_DIR="docs/adr"
EXIT_CODE=0

echo "🔍 Validating architecture documentation SSOT..."
echo ""

# Check 1: ARCHITECTURE.md exists
if [ ! -f "$ARCHITECTURE_FILE" ]; then
    echo "❌ ARCHITECTURE.md not found at $ARCHITECTURE_FILE"
    exit 1
fi

# Check 2: ARCHITECTURE.md doesn't contain decision language
# UNLESS it's marked with <!-- TODO: ADR --> comment
echo "Checking for decision language in ARCHITECTURE.md..."
DECISION_KEYWORDS=(
    "we chose"
    "we decided"
    "we selected"
    "because of"
    "the reason"
    "instead of"
    "alternative considered"
    "alternatives:"
    "options evaluated"
    "after evaluating"
)

FOUND_DECISION_LANGUAGE=0
for keyword in "${DECISION_KEYWORDS[@]}"; do
    # Find lines with decision language, excluding ADR references
    matches=$(grep -i -n "$keyword" "$ARCHITECTURE_FILE" | grep -v "ADR-" | grep -v "See \[" || true)

    if [ -n "$matches" ]; then
        # Check each match to see if it's near a TODO marker
        while IFS= read -r match; do
            line_num=$(echo "$match" | cut -d: -f1)

            # Check if there's a <!-- TODO: ADR --> within 5 lines before this line
            start_line=$((line_num - 5))
            if [ $start_line -lt 1 ]; then start_line=1; fi

            # Extract the range and check for TODO marker
            has_todo=0
            if sed -n "${start_line},${line_num}p" "$ARCHITECTURE_FILE" | grep -q "<!-- TODO: ADR"; then
                has_todo=1
            fi

            # If no TODO marker, report as violation
            if [ $has_todo -eq 0 ]; then
                if [ $FOUND_DECISION_LANGUAGE -eq 0 ]; then
                    echo "❌ Found decision language without TODO marker (belongs in ADRs):"
                    FOUND_DECISION_LANGUAGE=1
                    EXIT_CODE=1
                fi
                echo "   Line $line_num: '$keyword'"
                echo "      $(echo "$match" | cut -d: -f2-)"
            fi
        done <<< "$matches"
    fi
done

if [ $FOUND_DECISION_LANGUAGE -eq 0 ]; then
    echo "✅ No unmarked decision language (TODO markers found for known gaps)"
fi
echo ""

# Check 3: All ADRs are referenced in ARCHITECTURE.md
echo "Checking that all ADRs are referenced in ARCHITECTURE.md..."
MISSING_ADRS=()

# Find all ADR files (000-*.md pattern)
for adr_file in "$ADR_DIR"/[0-9][0-9][0-9]-*.md; do
    if [ -f "$adr_file" ]; then
        # Extract ADR number (000, 001, 002, etc.)
        adr_basename=$(basename "$adr_file")
        adr_number=$(echo "$adr_basename" | grep -o '^[0-9][0-9][0-9]')

        # Check if "ADR-000" or "[ADR-000]" or "(000-" appears in ARCHITECTURE.md
        if ! grep -q "ADR-$adr_number" "$ARCHITECTURE_FILE" && \
           ! grep -q "($adr_number-" "$ARCHITECTURE_FILE"; then
            MISSING_ADRS+=("ADR-$adr_number ($adr_basename)")
        fi
    fi
done

if [ ${#MISSING_ADRS[@]} -gt 0 ]; then
    echo "❌ The following ADRs are not referenced in ARCHITECTURE.md:"
    for missing in "${MISSING_ADRS[@]}"; do
        echo "   - $missing"
    done
    EXIT_CODE=1
else
    echo "✅ All ADRs are referenced in ARCHITECTURE.md"
fi
echo ""

# Check 4: ARCHITECTURE.md has required sections
echo "Checking for required sections in ARCHITECTURE.md..."
REQUIRED_SECTIONS=(
    "## Overview"
    "## Architectural Layers"
    "## Component Structure"
    "## How ADRs Work Together"
)

MISSING_SECTIONS=()
for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -q "$section" "$ARCHITECTURE_FILE"; then
        MISSING_SECTIONS+=("$section")
    fi
done

if [ ${#MISSING_SECTIONS[@]} -gt 0 ]; then
    echo "❌ Missing required sections in ARCHITECTURE.md:"
    for missing in "${MISSING_SECTIONS[@]}"; do
        echo "   - $missing"
    done
    EXIT_CODE=1
else
    echo "✅ All required sections present"
fi
echo ""

# Check 5: ARCHITECTURE.md links to ADR files correctly
echo "Checking ADR link format..."
# Links should be like [ADR-001](001-ultra-thin-command-pattern.md)
# Extract all ADR links and verify the files exist
BROKEN_LINKS=()

while IFS= read -r link; do
    # Extract filename from markdown link [text](filename.md)
    filename=$(echo "$link" | sed -n 's/.*(\([^)]*\.md\)).*/\1/p')
    if [ -n "$filename" ]; then
        full_path="$ADR_DIR/$filename"
        if [ ! -f "$full_path" ]; then
            BROKEN_LINKS+=("$filename")
        fi
    fi
done < <(grep -o '\[ADR-[0-9][0-9][0-9]\]([^)]*.md)' "$ARCHITECTURE_FILE" || true)

if [ ${#BROKEN_LINKS[@]} -gt 0 ]; then
    echo "❌ Broken ADR links in ARCHITECTURE.md:"
    for broken in "${BROKEN_LINKS[@]}"; do
        echo "   - $broken"
    done
    EXIT_CODE=1
else
    echo "✅ All ADR links are valid"
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Architecture documentation validation passed"
    echo ""
    echo "SSOT maintained:"
    echo "  • ARCHITECTURE.md contains structure (WHAT)"
    echo "  • ADRs contain decisions (WHY)"
    echo "  • No duplication detected"
else
    echo "❌ Architecture documentation validation failed"
    echo ""
    echo "Fix the issues above to maintain SSOT."
    echo ""
    echo "Guidelines:"
    echo "  • ARCHITECTURE.md should describe WHAT exists (structure, flow)"
    echo "  • ADRs should describe WHY (decisions, rationale, alternatives)"
    echo "  • Link to ADRs instead of duplicating decision text"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

exit $EXIT_CODE
