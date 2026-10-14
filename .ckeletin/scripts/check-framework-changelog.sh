#!/usr/bin/env bash
# Fail when framework code moves without a record in .ckeletin/CHANGELOG.md.
#
# .ckeletin/ is what `task ckeletin:update` checks out. The root CHANGELOG.md
# stays in the reference repository, so a derived CLI never sees it. This check
# keeps the framework changelog as the record of that checkout.
#
# Rules:
#   1. .ckeletin/VERSION must appear as a ## [X.Y.Z] heading in the changelog.
#   2. Since the last commit that touched .ckeletin/VERSION or
#      .ckeletin/CHANGELOG.md, any other change under .ckeletin/ (except
#      .ckeletin/docs/) must arrive with a changelog edit.
# A version bump that is not committed yet does not replay earlier history:
# only the files that differ from that last changelog/version commit count.

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/check-output.sh
source "${SCRIPT_DIR}/lib/check-output.sh"

check_header "Checking the framework changelog"

if [ ! -f .ckeletin/VERSION ]; then
	check_failure \
		"Framework version file is missing" \
		".ckeletin/VERSION" \
		"Restore .ckeletin/VERSION. It is the label task ckeletin:update prints."
	exit 1
fi

if [ ! -f .ckeletin/CHANGELOG.md ]; then
	check_failure \
		"Framework changelog is missing" \
		".ckeletin/CHANGELOG.md" \
		"Restore .ckeletin/CHANGELOG.md. Derived CLIs read it after an update."
	exit 1
fi

version=$(tr -d '[:space:]' < .ckeletin/VERSION)
if [ -z "$version" ]; then
	check_failure \
		"Framework version is empty" \
		".ckeletin/VERSION" \
		"Write the framework version, for example 0.2.0."
	exit 1
fi

if ! grep -q "^## \\[${version}\\]" .ckeletin/CHANGELOG.md; then
	check_failure \
		"Framework changelog has no section for v${version}" \
		"Expected a heading: ## [${version}]" \
		"Add the section to .ckeletin/CHANGELOG.md. The update prints the bullets between the old version and this one."
	exit 1
fi

if ! git rev-parse --is-inside-work-tree > /dev/null 2>&1 || ! git rev-parse HEAD > /dev/null 2>&1; then
	check_success "Framework changelog records v${version} (no commits to compare yet)"
	exit 0
fi

base=$(git log -1 --format=%H -- .ckeletin/VERSION .ckeletin/CHANGELOG.md 2> /dev/null || true)
if [ -z "$base" ]; then
	check_failure \
		"Framework version has no git history" \
		".ckeletin/VERSION and .ckeletin/CHANGELOG.md" \
		"Commit the version and the changelog before changing other framework files."
	exit 1
fi

changelog_changed=0
code_changes=""
seen="
"
while IFS= read -r path; do
	[ -z "$path" ] && continue
	case "
$seen" in
	*"
$path
"*) continue ;;
	esac
	seen="${seen}${path}
"
	case "$path" in
	.ckeletin/docs | .ckeletin/docs/*) continue ;;
	.ckeletin/CHANGELOG.md) changelog_changed=1 ;;
	.ckeletin/VERSION) ;;
	.ckeletin/*)
		code_changes="${code_changes}${path}
"
		;;
	esac
done << EOF
$(
	git diff --name-only "$base" -- .ckeletin
	git diff --name-only --cached "$base" -- .ckeletin
	git ls-files --others --exclude-standard -- .ckeletin
)
EOF

if [ -n "$code_changes" ] && [ "$changelog_changed" -eq 0 ]; then
	check_failure \
		"Framework files changed without a .ckeletin/CHANGELOG.md edit" \
		"$code_changes" \
		"Record the change under [Unreleased] in .ckeletin/CHANGELOG.md, or bump .ckeletin/VERSION and add that version's section. Docs under .ckeletin/docs/ are exempt."
	exit 1
fi

check_success "Framework changelog records v${version}"
exit 0
