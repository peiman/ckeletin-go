#!/usr/bin/env bash
# .ckeletin/scripts/check-framework-updates.sh
#
# Compare .ckeletin/VERSION with upstream. An update exists only when the
# upstream framework version is a newer semver. A commit that does not change
# that version is not an update. A local version ahead of upstream is not one
# either.
#
#   (no args)  notice for `task check`: throttled, skipped in CI, opt-out,
#              always exits 0, silent unless an update is available
#   --check    explicit text check; fetches every time; always exits 0
#   --json     explicit JSON check; fetches every time; always exits 0
#
# When a comparison is made the JSON object has current, latest, and
# update_available. When the check does not apply it has applicable: false.

set -o pipefail

mode=notice
case "${1:-}" in
    --check) mode=check ;;
    --json) mode=json ;;
    "") mode=notice ;;
    *) mode=notice ;;
esac

json_escape() {
    # Emit a JSON string, or null when the value is empty or not safe to embed.
    local v=$1
    if [ -z "$v" ]; then
        printf 'null'
        return
    fi
    case "$v" in
        *\"* | *\\* | *[[:cntrl:]]*) printf 'null' ;;
        *) printf '"%s"' "$v" ;;
    esac
}

# Sets PARSED_MAJ, PARSED_MIN, PARSED_PAT, and PARSED_PRE. Bash 3.2 has no namerefs.
semver_parse() {
    local raw=$1
    raw=${raw#v}
    raw=${raw#V}
    local core=${raw%%+*}
    PARSED_PRE=""
    if [[ "$core" == *-* ]]; then
        PARSED_PRE=${core#*-}
        core=${core%%-*}
    fi
    if [[ ! "$core" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        return 1
    fi
    PARSED_MAJ=${BASH_REMATCH[1]}
    PARSED_MIN=${BASH_REMATCH[2]}
    PARSED_PAT=${BASH_REMATCH[3]}
    return 0
}

# Return 0 when $1 (upstream) is a strictly newer semver than $2 (local).
semver_is_newer() {
    semver_parse "$1" || return 1
    local up_maj=$PARSED_MAJ up_min=$PARSED_MIN up_pat=$PARSED_PAT up_pre=$PARSED_PRE
    semver_parse "$2" || return 1
    local loc_maj=$PARSED_MAJ loc_min=$PARSED_MIN loc_pat=$PARSED_PAT loc_pre=$PARSED_PRE
    if ((10#$up_maj > 10#$loc_maj)); then return 0; fi
    if ((10#$up_maj < 10#$loc_maj)); then return 1; fi
    if ((10#$up_min > 10#$loc_min)); then return 0; fi
    if ((10#$up_min < 10#$loc_min)); then return 1; fi
    if ((10#$up_pat > 10#$loc_pat)); then return 0; fi
    if ((10#$up_pat < 10#$loc_pat)); then return 1; fi
    # Same core: a release is newer than its pre-release. Two pre-releases are not an update.
    if [ -z "$up_pre" ] && [ -n "$loc_pre" ]; then
        return 0
    fi
    return 1
}

read_local_version() {
    tr -d '[:space:]' < .ckeletin/VERSION 2> /dev/null || true
}

emit_json_comparison() {
    local current=$1 latest=$2 available=$3
    printf '{"current":%s,"latest":%s,"update_available":%s}\n' \
        "$(json_escape "$current")" "$(json_escape "$latest")" "$available"
}

finish_exempt() {
    local reason=$1
    local text=$2
    if [ "$mode" = "json" ]; then
        printf '{"applicable":false,"reason":"%s"}\n' "$reason"
    elif [ "$mode" = "check" ]; then
        printf '%s\n' "$text"
    fi
    exit 0
}

if [ "$mode" = "notice" ]; then
    if [ -n "${CI:-}" ]; then exit 0; fi
    if [ "${CKELETIN_SKIP_UPDATE_CHECK:-}" = "1" ]; then exit 0; fi
fi

# Skip in the upstream repo itself. The module path is split so scaffold-init
# does not rewrite this reference.
CURRENT_MODULE=$(head -1 go.mod 2> /dev/null | awk '{print $2}')
UPSTREAM_MODULE="github.com/peiman""/ckeletin-go"
if [ -z "$CURRENT_MODULE" ] || [ "$CURRENT_MODULE" = "$UPSTREAM_MODULE" ]; then
    finish_exempt "upstream repository" "This is the upstream repository — framework update check does not apply here."
fi

if ! git remote get-url ckeletin-upstream > /dev/null 2>&1; then
    if [ "$mode" = "json" ]; then
        printf '{"applicable":false,"reason":"no_upstream_remote"}\n'
    elif [ "$mode" = "check" ]; then
        echo "No upstream remote configured (run task ckeletin:update to set one up)."
    fi
    exit 0
fi

should_fetch=1
if [ "$mode" = "notice" ]; then
    CACHE_DIR=".ckeletin/cache"
    STAMP="${CACHE_DIR}/.last-update-check"
    TTL_SECONDS="${CKELETIN_UPDATE_CHECK_TTL:-86400}"
    case "$TTL_SECONDS" in
        '' | *[!0-9]*) TTL_SECONDS=86400 ;;
    esac
    NOW=$(date +%s 2> /dev/null || echo 0)
    if [ -f "$STAMP" ]; then
        LAST=$(cat "$STAMP" 2> /dev/null || echo 0)
        case "$LAST" in
            '' | *[!0-9]*) LAST=0 ;;
        esac
        if [ "$NOW" -ne 0 ] && [ $((NOW - LAST)) -lt "$TTL_SECONDS" ]; then
            should_fetch=0
        fi
    fi
fi

fetched=0
if [ "$should_fetch" -eq 1 ]; then
    TIMEOUT=""
    if command -v timeout > /dev/null 2>&1; then
        TIMEOUT="timeout 5"
    elif command -v gtimeout > /dev/null 2>&1; then
        TIMEOUT="gtimeout 5"
    fi
    # shellcheck disable=SC2086
    if $TIMEOUT git fetch ckeletin-upstream main --quiet > /dev/null 2>&1; then
        fetched=1
    fi
    if [ "$mode" = "notice" ]; then
        mkdir -p "$CACHE_DIR" 2> /dev/null || true
        echo "${NOW:-0}" > "$STAMP" 2> /dev/null || true
    fi
elif git rev-parse --verify --quiet ckeletin-upstream/main > /dev/null 2>&1; then
    fetched=1
fi

LOCAL=$(read_local_version)
if [ "$fetched" -ne 1 ]; then
    if [ "$mode" = "json" ]; then
        emit_json_comparison "$LOCAL" "" false
    elif [ "$mode" = "check" ]; then
        echo "Framework update status is unknown (upstream unreachable)."
    fi
    exit 0
fi

REMOTE=$(git show "ckeletin-upstream/main:.ckeletin/VERSION" 2> /dev/null | tr -d '[:space:]' || true)
available=false
if semver_is_newer "$REMOTE" "$LOCAL"; then
    available=true
fi

if [ "$mode" = "json" ]; then
    emit_json_comparison "$LOCAL" "$REMOTE" "$available"
    exit 0
fi

if [ "$available" = "true" ]; then
    echo ""
    echo "ℹ️  Framework update available: v${LOCAL} → v${REMOTE}."
    echo "   Run: task ckeletin:update        (preview first: task ckeletin:update:dry-run)"
    exit 0
fi

if [ "$mode" = "check" ]; then
    if [ "$LOCAL" = "$REMOTE" ]; then
        echo "Framework is up to date (v${LOCAL})."
    else
        echo "Framework v${LOCAL} is not behind upstream v${REMOTE}."
    fi
fi
exit 0
