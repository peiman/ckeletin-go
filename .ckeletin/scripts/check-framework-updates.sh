#!/usr/bin/env bash
# .ckeletin/scripts/check-framework-updates.sh
#
# Non-fatal nudge shown as a footer in `task check` when the ckeletin framework
# (.ckeletin/) is behind upstream, so downstream projects know to run
# `task ckeletin:update`. The whole value of the framework is "scaffold once,
# stay current via update" — this makes "you're behind" visible.
#
# Designed to NEVER fail the build and NEVER slow CI or offline runs:
#   - skips in CI ($CI) and when opted out (CKELETIN_SKIP_UPDATE_CHECK=1)
#   - skips in the upstream repo itself and when no upstream remote is configured
#   - throttles the network fetch to at most once per TTL (default 24h)
#   - time-boxes the fetch and swallows all errors
#   - always exits 0

set -o pipefail

# 1. Never run in CI — keep pipelines fast, offline-safe and network-free.
if [ -n "${CI:-}" ]; then exit 0; fi

# 2. Honour an explicit opt-out.
if [ "${CKELETIN_SKIP_UPDATE_CHECK:-}" = "1" ]; then exit 0; fi

# 3. Skip in the upstream repo itself — there is nothing to update from.
#    The module path is split so scaffold-init does not rewrite this reference.
CURRENT_MODULE=$(head -1 go.mod 2> /dev/null | awk '{print $2}')
UPSTREAM_MODULE="github.com/peiman""/ckeletin-go"
if [ -z "$CURRENT_MODULE" ] || [ "$CURRENT_MODULE" = "$UPSTREAM_MODULE" ]; then exit 0; fi

# 4. Need the upstream remote (set up by `task ckeletin:update`).
if ! git remote get-url ckeletin-upstream > /dev/null 2>&1; then exit 0; fi

# 5. Throttle: only hit the network at most once per TTL window.
CACHE_DIR=".ckeletin/cache"
STAMP="${CACHE_DIR}/.last-update-check"
TTL_SECONDS="${CKELETIN_UPDATE_CHECK_TTL:-86400}"
case "$TTL_SECONDS" in
    '' | *[!0-9]*) TTL_SECONDS=86400 ;;
esac
NOW=$(date +%s 2> /dev/null || echo 0)

should_fetch=1
if [ -f "$STAMP" ]; then
    LAST=$(cat "$STAMP" 2> /dev/null || echo 0)
    case "$LAST" in
        '' | *[!0-9]*) LAST=0 ;;
    esac
    if [ "$NOW" -ne 0 ] && [ $((NOW - LAST)) -lt "$TTL_SECONDS" ]; then
        should_fetch=0
    fi
fi

if [ "$should_fetch" -eq 1 ]; then
    # Pick a timeout wrapper if available (macOS often lacks `timeout`).
    TIMEOUT=""
    if command -v timeout > /dev/null 2>&1; then
        TIMEOUT="timeout 5"
    elif command -v gtimeout > /dev/null 2>&1; then
        TIMEOUT="gtimeout 5"
    fi
    # shellcheck disable=SC2086
    $TIMEOUT git fetch ckeletin-upstream main --quiet > /dev/null 2>&1 || true
    mkdir -p "$CACHE_DIR" 2> /dev/null || true
    echo "$NOW" > "$STAMP" 2> /dev/null || true
fi

# 6. Compare: how many upstream commits touch .ckeletin/ that we don't have yet.
BEHIND=$(git rev-list --count HEAD..ckeletin-upstream/main -- .ckeletin/ 2> /dev/null || echo 0)
case "$BEHIND" in
    '' | *[!0-9]*) BEHIND=0 ;;
esac

if [ "$BEHIND" -gt 0 ]; then
    echo ""
    echo "ℹ️  Framework update available: .ckeletin/ is ${BEHIND} commit(s) behind upstream."
    echo "   Run: task ckeletin:update        (preview first: task ckeletin:update:dry-run)"
fi

exit 0
