#!/usr/bin/env bash
# .ckeletin/scripts/check-release-drift.sh
#
# Detects drift in RELEASE INFRASTRUCTURE that `task ckeletin:update` does NOT
# sync (.goreleaser.yml and .github/workflows/* live in the project root, not
# under .ckeletin/). This drift is otherwise invisible until the first `git tag`
# triggers GoReleaser and it fails — the least-reversible, most-expensive moment.
# See issue #99.
#
# Checks (each is independent; all run, results are aggregated):
#   1. GoReleaser config validity + deprecated properties  (`goreleaser check`)
#   2. Stale CI secret token names in workflows            (e.g. CKELETIN_GITHUB_TOKEN)
#   3. Outdated cosign signing args in .goreleaser.yml     (--output-signature/-certificate)
#   4. Orphan v* tags not reachable from HEAD              (inherited from a non-fresh clone)
#
# Exit code: 0 when clean, 1 when any drift is found. Run this BEFORE tagging a release.

# NOTE: intentionally no `set -e` — every check must run so we can aggregate.
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Checking release infrastructure for drift"
echo ""

ISSUES=0
note_issue() { ISSUES=$((ISSUES + 1)); }

# Resolve the GoReleaser config path (either extension).
GORELEASER_FILE=""
if [ -f .goreleaser.yml ]; then
    GORELEASER_FILE=".goreleaser.yml"
elif [ -f .goreleaser.yaml ]; then
    GORELEASER_FILE=".goreleaser.yaml"
fi

# --- 1. GoReleaser config validity + deprecated properties -------------------
if [ -n "$GORELEASER_FILE" ]; then
    if ! command -v goreleaser > /dev/null 2>&1; then
        echo "  ⏭️  goreleaser not installed — skipped config validity check"
        echo "      Install: brew install goreleaser   (or https://goreleaser.com/install)"
    elif ! git remote 2> /dev/null | grep -q .; then
        # `goreleaser check` resolves SCM refs and needs a remote; skip without one
        # (e.g. a freshly-initialised project). Deprecations still surface in CI.
        echo "  ⏭️  no git remote — skipped goreleaser config check (needs a remote)"
    else
        if ! GR_OUT=$(goreleaser check 2>&1); then
            # Non-zero exit: removed/unknown properties or invalid config.
            check_failure \
                "GoReleaser config is invalid (removed or unknown properties)" \
                "$GR_OUT" \
                "Run 'goreleaser check' and fix the reported properties"$'\n'"See https://goreleaser.com/deprecations"
            note_issue
        elif echo "$GR_OUT" | grep -qiE "deprecat|phased out|will be removed"; then
            # `goreleaser check` exits 0 on still-valid-but-deprecated properties;
            # surface them so they are fixed before they are removed in a future release.
            check_failure \
                "GoReleaser config uses deprecated properties" \
                "$GR_OUT" \
                "Update the deprecated properties before they are removed"$'\n'"See https://goreleaser.com/deprecations"
            note_issue
        else
            check_success "GoReleaser config valid — no deprecations ($GORELEASER_FILE)"
        fi
    fi
fi

# --- 2. Stale CI secret token names ------------------------------------------
# The framework moved off a project-specific CKELETIN_GITHUB_TOKEN to the default
# GITHUB_TOKEN (with an optional RELEASE_GITHUB_TOKEN override). A stale copy is
# empty in a downstream repo and breaks GoReleaser with "missing GITHUB_TOKEN".
if [ -d .github/workflows ]; then
    STALE_TOKENS=$(grep -rnE "secrets\.CKELETIN_GITHUB_TOKEN" .github/workflows/ 2> /dev/null || true)
    if [ -n "$STALE_TOKENS" ]; then
        check_failure \
            "Stale CI secret token name (CKELETIN_GITHUB_TOKEN)" \
            "$STALE_TOKENS" \
            'Replace with: ${{ secrets.RELEASE_GITHUB_TOKEN || secrets.GITHUB_TOKEN }}'
        note_issue
    fi
fi

# --- 3. Outdated cosign signing args -----------------------------------------
# Newer cosign (after a cosign-installer bump) ignores --output-signature /
# --output-certificate and fails with "create bundle file: ... no such file".
if [ -n "$GORELEASER_FILE" ]; then
    # Exclude YAML comment-only lines so a "# replaced --output-signature" note
    # in a downstream config doesn't trigger a false positive.
    OLD_COSIGN=$(grep -nE -- "--output-signature|--output-certificate" "$GORELEASER_FILE" 2> /dev/null \
        | grep -vE '^[0-9]+:[[:space:]]*#' || true)
    if [ -n "$OLD_COSIGN" ]; then
        check_failure \
            "Outdated cosign signing flags in $GORELEASER_FILE" \
            "$OLD_COSIGN" \
            "Newer cosign ignores these flags. Use: --bundle=\${signature} --new-bundle-format"
        note_issue
    fi
fi

# --- 4. Orphan v* tags not reachable from HEAD -------------------------------
# A non-fresh clone (scaffolded before `init` ran `rm -rf .git`) inherits the
# framework's old tags (v0.0.1..). They silently hijack `git tag vX.Y.Z` and
# push the wrong commit. A fresh `init` history has none.
if git rev-parse --git-dir > /dev/null 2>&1; then
    if [ "$(git rev-parse --is-shallow-repository 2> /dev/null)" = "true" ]; then
        echo "  ⏭️  shallow clone — skipped orphan-tag check (needs full history)"
    else
        ORPHANS=""
        while IFS= read -r tag; do
            [ -z "$tag" ] && continue
            if ! git merge-base --is-ancestor "$tag" HEAD 2> /dev/null; then
                ORPHANS="${ORPHANS}${tag}"$'\n'
            fi
        done < <(git tag -l 'v*')
        ORPHANS="${ORPHANS%$'\n'}" # drop trailing newline for clean output
        if [ -n "$ORPHANS" ]; then
            check_failure \
                "Inherited tags not reachable from HEAD (orphan tags)" \
                "$ORPHANS" \
                "These likely came from a non-fresh clone and can hijack 'git tag'."$'\n'"Remove each with: git tag -d <tag>"
            note_issue
        fi
    fi
fi

echo ""
if [ "$ISSUES" -eq 0 ]; then
    check_success "No release-infrastructure drift detected"
    exit 0
fi

echo "❌ Release-infrastructure drift detected (${ISSUES} issue(s)) — fix before tagging a release"
exit 1
