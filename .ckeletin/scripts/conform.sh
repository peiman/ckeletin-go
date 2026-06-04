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

# ── Require yq, and fail fast if the mapping is not valid YAML ───
# conform.sh parses the mapping with yq (one consistent YAML parser). An
# unparseable mapping — e.g. a check string with an invalid backslash escape —
# must fail the build, and therefore the release gate (CKSPEC-ENF-009), rather
# than be silently tolerated by a lenient text scan. This gate runs first so a
# broken mapping exits before any checks (including the conformance test suite),
# which keeps it cheap and recursion-free.
if ! command -v yq >/dev/null 2>&1; then
    echo "FAILED — 'yq' is required to parse $MAPPING_FILE but was not found on PATH."
    echo "  Install yq (https://github.com/mikefarah/yq) and re-run."
    exit 1
fi
# Must be mikefarah/yq v4 (Go): the parser below uses strenv(), which the Python
# yq (kislyuk) lacks. Probe the capability so a wrong variant fails here with a
# clear message instead of a cryptic "unknown function strenv" mid-run.
if ! printf 'a: 1\n' | probe=x yq '.a = strenv(probe)' >/dev/null 2>&1; then
    echo "FAILED — the 'yq' on PATH is not mikefarah/yq v4 (strenv unsupported)."
    echo "  conform.sh needs mikefarah/yq: https://github.com/mikefarah/yq"
    exit 1
fi
if ! yq '.' "$MAPPING_FILE" >/dev/null 2>&1; then
    echo "FAILED — $MAPPING_FILE is not valid YAML (yq could not parse it)."
    echo "  Run 'yq . $MAPPING_FILE' to see the parse error."
    echo "  Tip: a shell command containing regex backslashes must be written as a"
    echo "       literal block scalar (- |-) so it is taken verbatim — see the file header."
    exit 1
fi

# ── Parse helpers (all reads go through yq — one consistent parser) ──
# Requirement IDs contain hyphens, so paths use strenv() to pass the id/field
# in as data rather than interpolating untrusted text into the yq expression.

get_spec_version() {
    yq '.spec_version // ""' "$MAPPING_FILE"
}

get_requirement_ids() {
    yq '.requirements | keys | .[]' "$MAPPING_FILE"
}

# Get a scalar field (title, status, enforcement_level) from a requirement.
get_field() {
    local req_id="$1" field="$2"
    req="$req_id" field="$field" \
        yq '.requirements[strenv(req)][strenv(field)] // ""' "$MAPPING_FILE"
}

# Get array items (checks or violation_tests); empty/missing → no output.
get_array_items() {
    local req_id="$1" field="$2"
    req="$req_id" field="$field" \
        yq '(.requirements[strenv(req)][strenv(field)] // [])[]' "$MAPPING_FILE"
}

# Print "true" if a requirement declares a violation_evidence block, else "false".
has_violation_evidence() {
    local req_id="$1"
    req="$req_id" yq '.requirements[strenv(req)] | has("violation_evidence")' "$MAPPING_FILE"
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
TOTAL=$(yq '.requirements | keys | length' "$MAPPING_FILE")

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

    # Guard: if python3 failed or JSON was malformed, EXPECTED_IDS is empty
    if [[ -z "$EXPECTED_IDS" ]]; then
        echo "FAILED — could not parse requirement IDs from spec JSON (python3 error or malformed data)."
        exit 1
    fi

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

# ── Report sync guard: the committed conformance-report.json must match what
# gen-conformance-report.sh produces from the mapping (mirrors the spec repo's
# requirements.json sync pattern; P9). A stale report would publish wrong status,
# so drift fails the build — and the release gate — here, before the checks run
# (fast, recursion-free), prompting a regenerate. The spec repo aggregates this
# published report instead of hand-authoring conformance/ckeletin-go.yaml.
# Runs AFTER the completeness check so a mapping that drops a requirement fails
# as "unmapped" (ENF-005) rather than as report drift.
REPORT_FILE="conformance-report.json"
GEN_SCRIPT=".ckeletin/scripts/gen-conformance-report.sh"
if [[ ! -f "$REPORT_FILE" ]]; then
    echo "FAILED — $REPORT_FILE is missing. Run 'task generate:conformance-report' and commit it."
    exit 1
fi
if ! diff <(bash "$GEN_SCRIPT" "$MAPPING_FILE") "$REPORT_FILE" >/dev/null 2>&1; then
    echo "FAILED — $REPORT_FILE is out of sync with $MAPPING_FILE."
    echo "  Run 'task generate:conformance-report' and commit the result."
    exit 1
fi
echo "Report sync: $REPORT_FILE matches the mapping (PASS)"
echo ""

# ── ENF-008 drift guard: machine-derivable facts in evidence must match reality.
# (A hand-typed count in prose silently rots — e.g. "35" after a 36th requirement.)
DRIFT=$(yq '.requirements[].evidence // ""' "$MAPPING_FILE" | grep -oE 'all [0-9]+ requirement' | grep -oE '[0-9]+' | head -1 || true)
if [[ -n "$DRIFT" && "$DRIFT" != "$TOTAL" ]]; then
    echo "evidence drift: prose claims '$DRIFT requirement IDs' but $TOTAL are mapped (ENF-008)" >> "$FAIL_FILE"
fi

# ENF-008 anchoring tallies: machine-enforced vs declared-analysis.
ENF008_AUTOMATED=0
ENF008_ANALYSIS=0

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

    # ── ENF-008: every met requirement must be machine-anchored (a check or a
    # violation test) OR a declared honor-system with a written analysis
    # (violation_evidence). No silent prose-only "met". ──
    if [[ "$status" == "met" ]]; then
        a_checks=$(get_array_items "$req_id" "checks")
        a_vtests=$(get_array_items "$req_id" "violation_tests")
        a_vevid=$(has_violation_evidence "$req_id")
        if [[ -n "$a_checks" || -n "$a_vtests" ]]; then
            ENF008_AUTOMATED=$((ENF008_AUTOMATED + 1))
        elif [[ "$enforcement" == "honor-system" && "$a_vevid" == "true" ]]; then
            ENF008_ANALYSIS=$((ENF008_ANALYSIS + 1))
        else
            echo "$req_id ($title): met but unanchored — add a check/violation_test, or declare honor-system + violation_evidence (ENF-008)" >> "$FAIL_FILE"
        fi
    fi

    # ── ENF-006: Check proof exists for claims above honor-system ──
    # Accepts either violation_tests OR violation_evidence (spec v0.4.0+)
    if [[ "$enforcement" != "honor-system" && "$enforcement" != "" ]]; then
        vtests=$(get_array_items "$req_id" "violation_tests")
        # "true" if this requirement declares a violation_evidence block.
        vevidence=$(has_violation_evidence "$req_id")

        if [[ -z "$vtests" && "$vevidence" != "true" ]]; then
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
            # Validate check command starts with an allowed prefix
            case "$check_cmd" in
                task\ *|test\ *|grep\ *|go\ *|"!"\ grep\ *|\!\ grep\ *)
                    ;; # allowed
                *)
                    echo "REJECTED"
                    echo "$req_id: check command rejected (not in allowlist): $check_cmd" >> "$FAIL_FILE"
                    continue
                    ;;
            esac
            printf "  %-20s %s ... " "$req_id" "$check_cmd"
            if bash -c "$check_cmd" > /dev/null 2>&1; then
                echo "ok"
            else
                echo "FAIL"
                echo "$req_id ($title): check FAILED: $check_cmd" >> "$FAIL_FILE"
            fi
        done
    fi
done

# ── Collect results ──────────────────────────────────────────────

MET=$(yq '[.requirements[] | select(.status == "met")] | length' "$MAPPING_FILE")
DEFERRED=$(yq '[.requirements[] | select(.status == "deferred")] | length' "$MAPPING_FILE")
PARTIAL=$(yq '[.requirements[] | select(.status == "partial")] | length' "$MAPPING_FILE")
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
echo "  Enforcement:   $ENF008_AUTOMATED automated, $ENF008_ANALYSIS analysis-with-evidence (ENF-008)"
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
