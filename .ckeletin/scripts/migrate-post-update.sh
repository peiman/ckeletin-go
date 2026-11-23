#!/bin/bash
# Post-update migrations for task ckeletin:update
# Each migration checks its precondition before running, so this script
# is safe to run multiple times (idempotent).

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Printed when a project Taskfile is missing generate:command, or still runs
# the generator itself. The vars are required: a bare `task:` forwarding drops
# name= and CLI_ARGS, which is how `task generate:command name=<name>` broke.
print_generate_command_forwarding() {
	cat << 'EOF'
  generate:command:
    desc: Generate a new command with full pattern (cmd + internal + tests + config)
    cmds:
      - task: ckeletin:generate:command
        vars:
          name: '{{.name}}'
          CLI_ARGS: '{{.CLI_ARGS}}'
EOF
}

# Migration 1: Remove stale public component from .go-arch-lint.yml
# When: pkg/ directory doesn't exist but .go-arch-lint.yml still references it
if [ ! -d "pkg" ] && [ -f ".go-arch-lint.yml" ] && grep -q "pkg/\*\*" .go-arch-lint.yml; then
    echo "   ✓ Removing stale pkg/** references from .go-arch-lint.yml"

    # Remove public component section, commonComponents entry, and deps entry
    # Uses a temp file for portable sed behavior across macOS and Linux
    tmpfile=$(mktemp)
    awk '
    /PUBLIC PACKAGES/      { skip=1; next }
    /pkg\/ contains/       { skip=1; next }
    /Do NOT import from/   { skip=1; next }
    /Can be imported by/   { skip=1; next }
    /Are independent of/   { skip=1; next }
    /See ADR-010/          { skip=1; next }
    /Cannot depend on any internal/ { skip=1; next }
    /Can use any external vendor/   { skip=1; next }
    /Public packages are completely/ { skip=1; next }
    /^  public:/           { skip=1; next }
    skip && /^[^ ]|^  [^ ]/ && !/^    / { skip=0 }
    skip                   { next }
    /- public/             { next }
    { print }
    ' .go-arch-lint.yml > "$tmpfile"
    mv "$tmpfile" .go-arch-lint.yml
fi

# Migration 2: JSON output types moved from internal/ui to .ckeletin/pkg/output
# When: internal/ui/json.go still exists (pre-refactoring state)
if [ -f "internal/ui/json.go" ] && grep -q "JSONEnvelope" internal/ui/json.go; then
    echo "   ✓ Removing internal/ui/json.go (moved to .ckeletin/pkg/output/)"
    rm -f internal/ui/json.go internal/ui/json_test.go
    echo "   ⚠ ACTION REQUIRED: Update imports in your code:"
    echo "     - Replace 'internal/ui' with '.ckeletin/pkg/output' for JSON types"
    echo "     - Affected: cmd/root.go, main.go, internal/check/executor.go"
    echo "     - Types moved: JSONEnvelope, JSONError, JSONResponder, IsJSONMode, SetOutputMode, etc."
fi

# Migration 3: Detect missing task forwardings in project Taskfile.yml
# When: .ckeletin/Taskfile.yml has tasks that the project Taskfile doesn't forward
# The expected-forwardings.txt file lists framework tasks that should have
# project-level aliases. This runs on every update to catch new tasks.
FORWARDINGS_FILE="$SCRIPT_DIR/expected-forwardings.txt"
if [ -f "$FORWARDINGS_FILE" ] && [ -f "Taskfile.yml" ]; then
    missing=()
    while IFS= read -r task_name; do
        # Skip empty lines and comments
        [[ -z "$task_name" || "$task_name" == \#* ]] && continue
        # Check if the project Taskfile has a forwarding for this task
        if ! grep -q "^  ${task_name}:" Taskfile.yml 2>/dev/null; then
            missing+=("$task_name")
        fi
    done < "$FORWARDINGS_FILE"

    if [ ${#missing[@]} -gt 0 ]; then
        echo ""
        echo "   ⚠ Missing task forwardings in Taskfile.yml:"
        echo "     The following framework tasks have no project-level alias."
        echo "     Add these to your Taskfile.yml under '# === Convenience Aliases ===':"
        echo ""
        for task_name in "${missing[@]}"; do
            if [ "$task_name" = "generate:command" ]; then
                print_generate_command_forwarding
                echo ""
                continue
            fi
            # Derive description from framework Taskfile. A missing framework
            # Taskfile must not abort the migration (pipefail + set -e).
            desc=$(grep -A1 "^  ${task_name}:" .ckeletin/Taskfile.yml 2> /dev/null | grep "desc:" | sed 's/.*desc: //' | head -1) || true
            echo "  ${task_name}:"
            if [ -n "$desc" ]; then
                echo "    desc: ${desc}"
            fi
            echo "    cmds: [task: ckeletin:${task_name}]"
            echo ""
        done
    fi
fi

# Migration 4: Ensure .go-arch-lint.yml includes all framework infrastructure packages
# When: Framework adds new packages (e.g., .ckeletin/pkg/output/) that aren't
# in the downstream project's infrastructure component
if [ -f ".go-arch-lint.yml" ]; then
    FRAMEWORK_INFRA_PKGS=(
        ".ckeletin/pkg/config/**"
        ".ckeletin/pkg/logger/**"
        ".ckeletin/pkg/progress/**"
        ".ckeletin/pkg/testutil/**"
    )
    missing_pkgs=()
    for pkg in "${FRAMEWORK_INFRA_PKGS[@]}"; do
        if ! grep -qF "$pkg" .go-arch-lint.yml; then
            missing_pkgs+=("$pkg")
        fi
    done
    if [ ${#missing_pkgs[@]} -gt 0 ]; then
        echo ""
        echo "   ⚠ Missing framework packages in .go-arch-lint.yml infrastructure component:"
        for pkg in "${missing_pkgs[@]}"; do
            echo "     - $pkg"
        done
        echo ""
        echo "   Add these to the 'infrastructure.in' section in .go-arch-lint.yml"
        echo "   to prevent 'not attached to any component' errors."
    fi
fi

# Migration 5: Exclude Claude Code worktrees from go-arch-lint
# When: .go-arch-lint.yml has an excludeFiles list but no .claude entry.
# Claude Code creates nested module checkouts under .claude/worktrees/;
# without the exclusion, stale copies there fail validate:layering with
# "not attached to any component" (scanner-exclusion SSOT:
# .ckeletin/scripts/lib/check-output.sh).
if [ -f ".go-arch-lint.yml" ] \
    && grep -q "^excludeFiles:" .go-arch-lint.yml \
    && ! grep -q '\\.claude' .go-arch-lint.yml; then
    echo "   ✓ Excluding .claude/ worktrees from .go-arch-lint.yml"
    tmpfile=$(mktemp)
    awk '
    { print }
    /^excludeFiles:/ {
        print "  # Claude Code worktrees are nested module checkouts, not part of the"
        print "  # module architecture (scanner-exclusion SSOT: .ckeletin/scripts/lib/check-output.sh)"
        print "  - /\\.claude/.*"
    }
    ' .go-arch-lint.yml > "$tmpfile" && mv "$tmpfile" .go-arch-lint.yml
fi

# Migration 6: generate:command must forward into the framework task.
# The task body is project-owned, so ckeletin:update will not replace an old
# copy that still runs `go run ./.ckeletin/scripts/generate-command/`. That copy
# drops name=. A missing task is reported by migration 3; this catches the
# stale body, which still has the `generate:command:` key.
if [ -f "Taskfile.yml" ] \
    && grep -q "^  generate:command:" Taskfile.yml \
    && ! grep -q "ckeletin:generate:command" Taskfile.yml; then
    echo ""
    echo "   ⚠ generate:command does not forward to ckeletin:generate:command"
    echo "     The local task still runs the generator itself, so name= is dropped."
    echo "     Replace it with:"
    echo ""
    print_generate_command_forwarding
    echo ""
fi

# Migration 7: pkg/output is its own component, not infrastructure.
# When: the output package exists and .go-arch-lint.yml has no output component.
# task ckeletin:update does not replace this file. The reference file already
# has the component, so this is a no-op there. A second run is a no-op too.
if [ -f ".go-arch-lint.yml" ] \
    && [ -d ".ckeletin/pkg/output" ] \
    && ! grep -qE '^  output:' .go-arch-lint.yml; then
    PYTHONUTF8=1 PYTHONIOENCODING=utf-8 python3 - << 'PY'
import re
import sys
from pathlib import Path

# Windows consoles default to cp1252. The status line uses a check mark.
for stream in (sys.stdout, sys.stderr):
    reconfigure = getattr(stream, "reconfigure", None)
    if reconfigure is not None:
        reconfigure(encoding="utf-8", errors="backslashreplace")

path = Path(".go-arch-lint.yml")
text = path.read_text()
if re.search(r"(?m)^  output:\s*$", text):
    sys.exit(0)

known = ("internal/ui/**", "internal/check/**", "internal/docs/**")
moved = [item for item in known if re.search(rf"(?m)^[ \t]+- {re.escape(item)}\s*$", text)]
for item in moved:
    text = re.sub(rf"(?m)^[ \t]+- {re.escape(item)}\s*\n", "", text)
text = re.sub(r"(?m)^[ \t]+- \.ckeletin/pkg/output/\*\*\s*\n", "", text)

output_component = (
    "  # Output writers. Not a common component, so business logic cannot import them.\n"
    "  output:\n"
    "    in:\n"
    "      - .ckeletin/pkg/output/**\n"
    "\n"
)
presentation_component = ""
if moved:
    presentation_component = (
        "  presentation:\n"
        "    in:\n"
        + "".join(f"      - {item}\n" for item in moved)
        + "\n"
    )

business = re.search(r"(?m)^  business:\s*$", text)
vendors = re.search(r"(?m)^vendors:\s*$", text)
if business and (vendors is None or business.start() < vendors.start()):
    text = text[: business.start()] + output_component + presentation_component + text[business.start() :]
else:
    sys.stderr.write("   ⚠ .go-arch-lint.yml has no business component; add an output component by hand\n")
    sys.exit(0)

def add_may_depend(text, component, dep):
    deps_at = text.find("\ndeps:\n")
    if deps_at < 0:
        return text
    match = re.search(
        rf"(?m)^  {component}:\n(?:    .*\n)*?    mayDependOn:\n",
        text[deps_at:],
    )
    if not match:
        return text
    start = deps_at + match.end()
    lines = text[start:].splitlines(keepends=True)
    existing = []
    for line in lines:
        if re.match(r"      - ", line):
            existing.append(line)
            continue
        break
    if any(re.search(rf"\b{dep}\b", line) for line in existing):
        return text
    return text[:start] + f"      - {dep}\n" + text[start:]

text = add_may_depend(text, "commands", "output")
if moved:
    text = add_may_depend(text, "commands", "presentation")

deps_at = text.find("\ndeps:\n")
infra = text.find("\n  infrastructure:\n", deps_at if deps_at >= 0 else 0)
blocks = (
    "  output:\n"
    "    mayDependOn:\n"
    "      - infrastructure\n"
    "    canUse:\n"
    "      - cobra\n"
    "      - pflag\n"
    "      - zerolog\n"
    "      - viper\n"
    "      - isatty\n"
    "\n"
)
if moved:
    blocks += (
        "  presentation:\n"
        "    mayDependOn:\n"
        "      - infrastructure\n"
        "      - output\n"
        "    canUse:\n"
        "      - zerolog\n"
        "      - viper\n"
        "      - isatty\n"
        "      - lipgloss\n"
        "      - bubbletea\n"
        "\n"
    )
if infra >= 0 and not re.search(r"(?m)^  output:\n    mayDependOn:", text):
    text = text[: infra + 1] + blocks + text[infra + 1 :]

path.write_text(text)
print("   ✓ Moved pkg/output into its own component")

# A business package that still imports the writers is left where it is.
# Moving it would hide the violation the layering check is there to catch.
importers = []
for go_file in Path(".").rglob("*.go"):
    rel = go_file.as_posix().lstrip("./")
    if rel.endswith("_test.go") or rel.startswith(("cmd/", ".ckeletin/pkg/output/", "test/")):
        continue
    if rel == "main.go":
        continue
    body = go_file.read_text(errors="replace")
    if '.ckeletin/pkg/output"' not in body:
        continue
    parent = str(go_file.parent).replace("\\", "/").lstrip("./")
    glob = parent + "/**"
    if glob in known or glob == ".ckeletin/pkg/output/**":
        continue
    importers.append(rel)
if importers:
    print("   ⚠ These files import pkg/output and were left in their component.")
    print("     Move a presentation package by hand, or report through progress.Reporter:")
    for rel in sorted(importers):
        print(f"     - {rel}")
PY
fi
