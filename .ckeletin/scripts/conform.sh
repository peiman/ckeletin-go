#!/usr/bin/env bash
# Conformance report generator for ckeletin-go.
# Reads conformance-mapping.yaml, runs checks, validates completeness,
# reports feedback signals.
#
# Implements:
#   CKSPEC-ENF-005 — mapping completeness (fail on unmapped requirements)
#   CKSPEC-ENF-006 — violation test verification
#   CKSPEC-ENF-007 — automatic feedback signals

set -euo pipefail

MAPPING_FILE="conformance-mapping.yaml"
FAIL_FILE=$(mktemp)
FEEDBACK_FILE=$(mktemp)
WARNING_FILE=$(mktemp)
trap 'rm -f "$FAIL_FILE" "$FEEDBACK_FILE" "$WARNING_FILE"' EXIT

# ── Parse helpers (YAML subset — no external deps) ──────────────

get_spec_version() {
    grep '^spec_version:' "$MAPPING_FILE" | head -1 | sed 's/spec_version: *"\(.*\)"/\1/'
}

get_requirement_ids() {
    grep '^  CKSPEC-' "$MAPPING_FILE" | sed 's/^ *\(CKSPEC-[A-Z]*-[0-9]*\):.*/\1/'
}

# Get a scalar field from a requirement block
get_field() {
    local req_id="$1" field="$2"
    awk -v req="$req_id" -v field="$field" '
        /^  [A-Z]/ && $0 ~ req":" { found=1; next }
        found && /^  [A-Z]/ { found=0 }
        found && $0 ~ "^    " field ":" {
            line=$0
            sub(/^[^:]*: */, "", line)
            gsub(/^ *"?/, "", line); gsub(/"? *$/, "", line)
            if (line != "" && line != ">") print line
            exit
        }
    ' "$MAPPING_FILE"
}

# Get array items (checks or violation_tests)
get_array_items() {
    local req_id="$1" field="$2"
    awk -v req="$req_id" -v field="$field" '
        /^  [A-Z]/ && $0 ~ req":" { found=1; next }
        found && /^  [A-Z]/ { found=0 }
        found && $0 ~ "^    " field ":" { in_array=1; next }
        in_array && /^    [a-z]/ { in_array=0; next }
        in_array && /^      - / {
            line=$0
            sub(/^ *- *"?/, "", line); sub(/"? *$/, "", line)
            if (line != "") print line
        }
    ' "$MAPPING_FILE"
}

# ── Main ────────────────────────────────────────────────────────

echo "ckeletin-go conformance check"
echo "================================"
echo ""

SPEC_VERSION=$(get_spec_version)

echo "Spec version (mapping): $SPEC_VERSION"
echo "Mapping file: $MAPPING_FILE"
echo ""

REQ_IDS=$(get_requirement_ids)
TOTAL=$(echo "$REQ_IDS" | wc -l | tr -d ' ')

echo "Requirements mapped: $TOTAL"
echo ""

# ── ENF-005: Completeness check ─────────────────────────────────
# Fetch the authoritative requirement list from the spec repo.
# Falls back to a hardcoded list if the fetch fails (offline mode).

SPEC_REPO="peiman/ckeletin"
SPEC_JSON_URL="https://raw.githubusercontent.com/${SPEC_REPO}/main/spec/requirements.json"
CACHE_FILE=".ckeletin/cache/requirements.json"
SPEC_JSON=""
EXPECTED_IDS=""
SPEC_LATEST_VERSION=""

# Try fetching from GitHub (silent, fast timeout)
if command -v curl &> /dev/null; then
    SPEC_JSON=$(curl -sfL --max-time 5 "$SPEC_JSON_URL" 2>/dev/null || true)
fi

if [[ -n "$SPEC_JSON" ]]; then
    # Cache the successful fetch for offline use
    mkdir -p "$(dirname "$CACHE_FILE")"
    echo "$SPEC_JSON" > "$CACHE_FILE"
    SOURCE="fetched from spec repo"
elif [[ -f "$CACHE_FILE" ]]; then
    # Fall back to last cached version
    SPEC_JSON=$(cat "$CACHE_FILE")
    SOURCE="cached (offline)"
fi

if [[ -n "$SPEC_JSON" ]]; then
    EXPECTED_IDS=$(echo "$SPEC_JSON" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for r in data['requirements']:
    print(r['id'])
" 2>/dev/null || true)
    SPEC_LATEST_VERSION=$(echo "$SPEC_JSON" | python3 -c "
import sys, json
print(json.load(sys.stdin)['spec_version'])
" 2>/dev/null || true)
    echo "Requirement list: ${SOURCE} (v${SPEC_LATEST_VERSION})"
else
    echo "Requirement list: no spec data available (fetch failed, no cache)"
    echo "FAILED — cannot validate completeness without requirement list."
    exit 1
fi

# Warn on spec version mismatch
if [[ -n "$SPEC_LATEST_VERSION" && "$SPEC_VERSION" != "$SPEC_LATEST_VERSION" ]]; then
    echo ""
    echo "⚠ SPEC VERSION MISMATCH"
    echo "  Mapping targets spec $SPEC_VERSION but latest spec is $SPEC_LATEST_VERSION"
    echo "  Update conformance-mapping.yaml to match the latest spec."
    echo ""
fi

echo ""

MISSING_COUNT=0
for expected in $EXPECTED_IDS; do
    [[ -z "$expected" ]] && continue
    if ! echo "$REQ_IDS" | grep -q "^${expected}$"; then
        echo "  MISSING: $expected"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done

if [[ $MISSING_COUNT -gt 0 ]]; then
    echo ""
    echo "FAILED — $MISSING_COUNT unmapped requirement(s) (CKSPEC-ENF-005 violation)."
    exit 1
fi

echo "Completeness: $TOTAL/$TOTAL requirements mapped (ENF-005: PASS)"
echo ""

# ── Run checks and validate ──────────────────────────────────────

echo "Running checks..."
echo ""

for req_id in $REQ_IDS; do
    title=$(get_field "$req_id" "title")
    status=$(get_field "$req_id" "status")
    enforcement=$(get_field "$req_id" "enforcement_level")

    if [[ "$status" == "deferred" ]]; then
        echo "$req_id ($title): deferred" >> "$WARNING_FILE"
    fi

    if [[ "$status" == "partial" ]]; then
        echo "$req_id ($title): partial" >> "$WARNING_FILE"
    fi

    # ── ENF-006: Check proof exists for claims above honor-system ──
    # Accepts either violation_tests OR violation_evidence (spec v0.4.0+)
    if [[ "$enforcement" != "honor-system" && "$enforcement" != "" ]]; then
        vtests=$(get_array_items "$req_id" "violation_tests")
        # Check if violation_evidence exists in this requirement's block
        # (multi-line field, so get_field may not capture it — use grep)
        vevidence=$(awk -v req="$req_id" '
            /^  [A-Z]/ && $0 ~ req":" { found=1; next }
            found && /^  [A-Z]/ { found=0 }
            found && /violation_evidence:/ { print "yes"; exit }
        ' "$MAPPING_FILE")

        if [[ -z "$vtests" && -z "$vevidence" ]]; then
            echo "$req_id: claims $enforcement but has no violation test or evidence" >> "$FEEDBACK_FILE"
        elif [[ -n "$vtests" ]]; then
            echo "$vtests" | while IFS= read -r vt; do
                # Strip test function reference (file.go::TestFunc -> file.go)
                vt_file="${vt%%::*}"
                if [[ -n "$vt_file" && ! -f "$vt_file" ]]; then
                    echo "$req_id: violation test file not found: $vt_file" >> "$FEEDBACK_FILE"
                fi
            done
        fi
        # violation_evidence is accepted at face value if it exists —
        # the file-path requirement is enforced by review, not tooling
    fi

    # ── Run automated checks ──
    checks=$(get_array_items "$req_id" "checks")
    if [[ -n "$checks" ]]; then
        echo "$checks" | while IFS= read -r check_cmd; do
            if [[ -z "$check_cmd" ]]; then continue; fi
            printf "  %-20s %s ... " "$req_id" "$check_cmd"
            if eval "$check_cmd" > /dev/null 2>&1; then
                echo "ok"
            else
                echo "FAIL"
                echo "$req_id ($title): check FAILED: $check_cmd" >> "$FAIL_FILE"
            fi
        done
    fi
done

# ── Collect results ──────────────────────────────────────────────

MET=$(grep -c 'status: met' "$MAPPING_FILE" || true)
DEFERRED=$(grep -c 'status: deferred' "$MAPPING_FILE" || true)
PARTIAL=$(grep -c 'status: partial' "$MAPPING_FILE" || true)
FAILED_CHECKS=0
if [[ -s "$FAIL_FILE" ]]; then
    FAILED_CHECKS=$(wc -l < "$FAIL_FILE" | tr -d ' ')
fi
WARNING_COUNT=0
if [[ -s "$WARNING_FILE" ]]; then
    WARNING_COUNT=$(wc -l < "$WARNING_FILE" | tr -d ' ')
fi
FEEDBACK_COUNT=0
if [[ -s "$FEEDBACK_FILE" ]]; then
    FEEDBACK_COUNT=$(wc -l < "$FEEDBACK_FILE" | tr -d ' ')
fi

echo ""
echo "── Results ──────────────────────────────────────────"
echo ""
echo "  Requirements:  $TOTAL total"
echo "  Met:           $MET"
echo "  Partial:       $PARTIAL"
echo "  Deferred:      $DEFERRED"
echo "  Failed checks: $FAILED_CHECKS"
echo ""

if [[ $WARNING_COUNT -gt 0 ]]; then
    echo "⚠ Warnings ($WARNING_COUNT):"
    sed 's/^/  - /' "$WARNING_FILE"
    echo ""
fi

if [[ $FAILED_CHECKS -gt 0 ]]; then
    echo "❌ Failed checks ($FAILED_CHECKS):"
    sed 's/^/  - /' "$FAIL_FILE"
    echo ""
fi

if [[ $FEEDBACK_COUNT -gt 0 ]]; then
    echo "📋 Feedback signals (ENF-007):"
    sed 's/^/  - /' "$FEEDBACK_FILE"
    echo ""
fi

# ── JSON summary ─────────────────────────────────────────────────

if [[ "${OUTPUT_JSON:-}" == "1" || "${1:-}" == "--json" ]]; then
    cat <<ENDJSON
{
  "implementation": "ckeletin-go",
  "spec_version": "$SPEC_VERSION",
  "total": $TOTAL,
  "met": $MET,
  "partial": $PARTIAL,
  "deferred": $DEFERRED,
  "failed_checks": $FAILED_CHECKS,
  "feedback_signals": $FEEDBACK_COUNT,
  "passed": $([ "$FAILED_CHECKS" -eq 0 ] && echo "true" || echo "false")
}
ENDJSON
fi

# ── Final verdict ────────────────────────────────────────────────

if [[ $FAILED_CHECKS -gt 0 ]]; then
    echo "FAILED — $FAILED_CHECKS check(s) did not pass."
    exit 1
fi

echo "PASSED — $MET/$TOTAL requirements met, $PARTIAL partial, $DEFERRED deferred."
if [[ $FEEDBACK_COUNT -gt 0 ]]; then
    echo "         $FEEDBACK_COUNT feedback signal(s) for spec review."
fi
