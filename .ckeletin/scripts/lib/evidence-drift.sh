#!/usr/bin/env bash
# ENF-008 guard: evidence must not hardcode the requirement count.
# Usage: source scripts/lib/evidence-drift.sh
#
# Sourced library: deliberately no `set -eo pipefail` here — shell options
# would leak into (or fight with) the sourcing script, which owns strict mode.

# evidence_count_drift <mapping-file> <mapped-total>
#
# Prints one failure line when any requirement's evidence contains
# "all <digits> requirement". A number that matches <mapped-total> still
# fails: the completeness line is the count, and a matching sentence goes
# stale the moment a requirement is added. Prints nothing when the prose
# does not restate the count.
evidence_count_drift() {
    local mapping="$1"
    local total="$2"
    local hit
    hit=$(yq '.requirements[].evidence // ""' "$mapping" | grep -oE 'all [0-9]+ requirement' | head -1 || true)
    if [[ -n "$hit" ]]; then
        printf "evidence drift: prose claims '%s' but the count is computed (%s mapped). Do not hardcode it (ENF-008)\n" "$hit" "$total"
    fi
}
