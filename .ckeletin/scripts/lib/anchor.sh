#!/usr/bin/env bash
# Violation-test anchor resolution for CKSPEC-ENF-008.
# Usage: source scripts/lib/anchor.sh
#
# Sourced library: deliberately no `set -eo pipefail` here — shell options
# would leak into (or fight with) the sourcing script, which owns strict mode.

# anchor_resolve <anchor> <fail_file> [req_id]
#
# ENF-008: an evidence anchor MUST resolve, not merely be present. For a
# violation_test `file::symbol` anchor that means the file must exist AND, when
# a symbol is named, a function with that symbol must be defined in it. A
# dangling anchor — a renamed/removed test still cited by its old name, or a
# garbled `file::` with no symbol — appends a "dangling anchor" line to
# <fail_file>; the caller treats a non-empty fail file as a hard failure.
# The optional receiver group also resolves method-style tests,
# func (s *Suite) TestX, not just free functions.
anchor_resolve() {
    local vt="$1" fail_file="$2" req_id="${3:-?}"
    [[ -z "$vt" ]] && return 0
    local vt_file="${vt%%::*}"
    local vt_symbol="${vt#*::}"
    if [[ -n "$vt_file" && ! -f "$vt_file" ]]; then
        echo "$req_id: dangling anchor — violation test file not found: $vt_file (ENF-008)" >> "$fail_file"
    elif [[ "$vt" == *::* && -z "$vt_symbol" ]]; then
        echo "$req_id: dangling anchor — empty symbol after '::' in '$vt' (ENF-008)" >> "$fail_file"
    elif [[ -n "$vt_symbol" && "$vt_symbol" != "$vt" ]] && \
         ! grep -qE "func[[:space:]]+(\([^)]*\)[[:space:]]+)?${vt_symbol}\b" "$vt_file"; then
        echo "$req_id: dangling anchor — symbol '$vt_symbol' not found in $vt_file (ENF-008)" >> "$fail_file"
    fi
}
