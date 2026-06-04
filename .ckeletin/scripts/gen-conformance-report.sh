#!/usr/bin/env bash
# Generate the machine-readable conformance report for ckeletin-go.
#
# Emits a DETERMINISTIC JSON projection of conformance-mapping.yaml to stdout:
# header (implementation, spec_version), summary totals, and per-requirement
# {status, enforcement_level, evidence, checks, violation_tests,
# violation_evidence}. The spec repo (peiman/ckeletin) aggregates this published
# report instead of hand-authoring conformance/ckeletin-go.yaml (CKSPEC SSOT).
#
# Determinism: yq extracts the DATA (semantics are YAML-spec-defined, so stable
# across yq versions) and python3 canonicalizes the FORMATTING (sorted keys,
# fixed indent). So the committed report is byte-stable regardless of which yq
# version produced it — which is what makes the `task conform` sync-check
# reliable in CI (yq 4.44.6) and locally (any mikefarah yq v4).
#
# There is intentionally NO timestamp in the output: a volatile field would
# break the sync-check. The spec-repo aggregator stamps the fetch date.
#
# Usage: gen-conformance-report.sh [mapping-file]   (default conformance-mapping.yaml)

set -euo pipefail

MAPPING_FILE="${1:-conformance-mapping.yaml}"

yq -o=json '
{
  "implementation": "ckeletin-go",
  "spec_version": .spec_version,
  "summary": {
    "total": (.requirements | length),
    "met": ([.requirements[] | select(.status == "met")] | length),
    "partial": ([.requirements[] | select(.status == "partial")] | length),
    "deferred": ([.requirements[] | select(.status == "deferred")] | length),
    "passed": (([.requirements[] | select(.status == "met")] | length) == (.requirements | length))
  },
  "requirements": (.requirements | map_values({
    "status": .status,
    "enforcement_level": .enforcement_level,
    "evidence": .evidence,
    "checks": (.checks // []),
    "violation_tests": (.violation_tests // []),
    "violation_evidence": .violation_evidence
  }))
}
' "$MAPPING_FILE" |
    python3 -c 'import json, sys; json.dump(json.load(sys.stdin), sys.stdout, indent=2, sort_keys=True, ensure_ascii=True); sys.stdout.write("\n")'
