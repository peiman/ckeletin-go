#!/usr/bin/env bash
# .ckeletin/scripts/validate-hook-forwardings.sh
#
# Enforces consistency between the framework git hooks and the forwarding SSOT.
#
# Every consumer extends .ckeletin/configs/lefthook.base.yml, where the framework
# tasks are namespaced `ckeletin:*`. A hook that runs a BARE `task <name>` only
# resolves downstream if `<name>` has a project-level forwarding — i.e. it is
# listed in expected-forwardings.txt (which migrate-post-update.sh ensures the
# consumer's Taskfile.yml provides). If a base hook references a bare task that
# is NOT in that list, the consumer's hook fails with "task does not exist" and
# blocks every push (issue #100). This check makes that disagreement loud here,
# in `task check`, instead of silently at a downstream's first push.
#
# A referenced task is OK when it is either:
#   - namespaced `ckeletin:*` (resolves in any consumer), or
#   - listed in expected-forwardings.txt (the consumer gets a forwarding).
#
# Exit 0 when consistent, 1 otherwise.

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

BASE_HOOKS=".ckeletin/configs/lefthook.base.yml"
# Overridable so tests can supply a controlled forwarding list (defaults to the
# real SSOT next to this script).
EXPECTED="${CKELETIN_EXPECTED_FORWARDINGS:-${SCRIPT_DIR}/expected-forwardings.txt}"

check_header "Validating lefthook hooks reference only forwarded tasks (issue #100)"

if [ ! -f "$BASE_HOOKS" ]; then
    echo "  ⏭️  $BASE_HOOKS not found — skipping"
    exit 0
fi
if [ ! -f "$EXPECTED" ]; then
    check_failure "expected-forwardings.txt not found" "$EXPECTED" "The forwarding SSOT is required to validate hook references"
    exit 1
fi

# Extract task names referenced by `task <name>` in the base hooks. Comments are
# stripped first (sed 's/#.*//'), and only `run:` directives are considered, so a
# task name appearing in a desc: or prose is not mistaken for a hook reference.
# `|| true`: an empty match must yield no references, not abort under `set -e`.
REFERENCED=$(sed 's/#.*//' "$BASE_HOOKS" | grep -E '^[[:space:]]*run:' | grep -oE 'task [a-z][a-zA-Z0-9:_-]*' | awk '{print $2}' | sort -u || true)

MISSING=""
for task in $REFERENCED; do
    case "$task" in
        ckeletin:*) continue ;; # namespaced — resolves in every consumer
    esac
    if ! grep -Fxq "$task" "$EXPECTED"; then
        MISSING="${MISSING}  ${task}"$'\n'
    fi
done
MISSING="${MISSING%$'\n'}"

if [ -n "$MISSING" ]; then
    check_failure \
        "Base hooks reference bare tasks with no forwarding" \
        "$MISSING" \
        "Each task a hook runs must resolve in a consumer. Either:"$'\n'"  - namespace it in lefthook.base.yml (task ckeletin:<name>), or"$'\n'"  - add <name> to .ckeletin/scripts/expected-forwardings.txt"
    exit 1
fi

check_success "All base-hook task references are namespaced or forwarded"
exit 0
