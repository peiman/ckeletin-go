#!/usr/bin/env bash
# Release the framework. Feature commits record notes under [Unreleased] and
# do not edit .ckeletin/VERSION. This command writes the next version, moves
# those notes under that heading, commits, and tags framework-vX.Y.Z.
# The tag is the release. Pass --push to publish the commit and the tag.
set -euo pipefail

bump="patch"
do_push=0
while [[ $# -gt 0 ]]; do
	case "$1" in
	--bump)
		bump="${2:-}"
		shift 2
		;;
	--push)
		do_push=1
		shift
		;;
	*)
		echo "Unknown argument: $1" >&2
		echo "Usage: release-framework.sh [--bump patch|minor|major] [--push]" >&2
		exit 1
		;;
	esac
done

case "$bump" in
patch | minor | major) ;;
*)
	echo "Bump must be patch, minor, or major." >&2
	exit 1
	;;
esac

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
	echo "Not a git repository." >&2
	exit 1
fi

branch="$(git branch --show-current)"
if [[ "$branch" != "main" && "${CKELETIN_RELEASE_ALLOW_BRANCH:-}" != "1" ]]; then
	echo "Release the framework from main. This branch is ${branch:-detached}." >&2
	exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
	echo "The working tree is not clean. Commit or stash before releasing." >&2
	exit 1
fi

module="$(awk 'NR==1 {print $2}' go.mod 2>/dev/null || true)"
if [[ "$module" != "github.com/peiman/ckeletin-go" && "${CKELETIN_RELEASE_ALLOW_BRANCH:-}" != "1" ]]; then
	echo "Only the ckeletin-go repository releases the framework." >&2
	exit 1
fi

if [[ ! -f .ckeletin/VERSION || ! -f .ckeletin/CHANGELOG.md ]]; then
	echo ".ckeletin/VERSION and .ckeletin/CHANGELOG.md are required." >&2
	exit 1
fi

current="$(tr -d '[:space:]' < .ckeletin/VERSION)"
if [[ ! "$current" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
	echo "VERSION is not X.Y.Z: $current" >&2
	exit 1
fi
major="${BASH_REMATCH[1]}"
minor="${BASH_REMATCH[2]}"
patch="${BASH_REMATCH[3]}"
case "$bump" in
patch) patch=$((patch + 1)) ;;
minor)
	minor=$((minor + 1))
	patch=0
	;;
major)
	major=$((major + 1))
	minor=0
	patch=0
	;;
esac
next="${major}.${minor}.${patch}"
today="$(date -u +%Y-%m-%d)"

python3 - "$next" "$today" <<'PY'
import pathlib, sys
version, today = sys.argv[1], sys.argv[2]
path = pathlib.Path(".ckeletin/CHANGELOG.md")
text = path.read_text()
start = text.find("## [Unreleased]")
if start < 0:
    sys.exit("CHANGELOG.md has no ## [Unreleased] section.")
rest = text[start + len("## [Unreleased]") :]
nxt = rest.find("\n## [")
if nxt < 0:
    body, tail = rest, ""
else:
    body, tail = rest[:nxt], rest[nxt:]
notes = body.strip("\n")
if not any(line.startswith("- ") for line in notes.splitlines()):
    sys.exit("Nothing to release. [Unreleased] has no changelog bullets.")
released = f"## [{version}] - {today}\n\n{notes}\n"
path.write_text(text[:start] + "## [Unreleased]\n\n" + released + tail.lstrip("\n"))
PY

printf '%s\n' "$next" > .ckeletin/VERSION

git add .ckeletin/VERSION .ckeletin/CHANGELOG.md
git commit -m "chore: release framework v${next}"
git tag -a "framework-v${next}" -m "Framework v${next}"

echo "Tagged framework-v${next} at $(git rev-parse --short HEAD)"
if [[ "$do_push" -eq 1 ]]; then
	git push origin HEAD
	git push origin "framework-v${next}"
fi
