#!/bin/bash
# PreToolUse(Bash) hook: block git commits that carry Claude Code attribution.
#
# Claude Code delivers the tool call as JSON on stdin. For the Bash tool the
# command string is at .tool_input.command (current schema; .parameters.command
# was an older shape, kept as a fallback). To BLOCK the tool call, the hook must
# exit 2 — its stderr is shown to Claude. exit 0 lets the command proceed.

TOOL_JSON=$(cat)

COMMAND=$(printf '%s' "$TOOL_JSON" |
	jq -r '.tool_input.command // .parameters.command // empty' 2>/dev/null || true)

if [[ "$COMMAND" == *"git commit"* ]]; then
	if printf '%s' "$COMMAND" | grep -qE 'Generated with \[Claude Code\]|Co-Authored-By: Claude'; then
		echo "❌ Git commit contains Claude Code attribution — remove it before committing:" >&2
		echo "  - 🤖 Generated with [Claude Code](https://claude.com/claude-code)" >&2
		echo "  - Co-Authored-By: Claude <noreply@anthropic.com>" >&2
		echo "Commit messages should contain only technical content (what changed and why)." >&2
		exit 2 # exit 2 blocks the tool call; stderr is fed back to Claude
	fi
fi

exit 0
