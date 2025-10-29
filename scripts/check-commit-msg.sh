#!/bin/bash
# Hook to prevent Claude Code attribution in git commits
# PreToolUse hooks receive JSON via stdin containing tool details

# Debug: Log what we receive (for troubleshooting)
DEBUG_LOG="/tmp/claude-hook-debug.log"
echo "=== Hook triggered at $(date) ===" >> "$DEBUG_LOG"

# Read JSON from stdin
TOOL_JSON=$(cat)
echo "Received JSON: $TOOL_JSON" >> "$DEBUG_LOG"

# Extract the command from the tool parameters
# For Bash tool: {"tool":"Bash","parameters":{"command":"git commit ...","description":"..."}}
COMMAND=$(echo "$TOOL_JSON" | jq -r '.parameters.command // empty' 2>/dev/null || echo "")
echo "Extracted command: $COMMAND" >> "$DEBUG_LOG"

# Check if this is a git commit command
if [[ "$COMMAND" == *"git commit"* ]]; then
    # Check for Claude attribution patterns
    if echo "$COMMAND" | grep -q "Generated with \[Claude Code\]" || \
       echo "$COMMAND" | grep -q "Co-Authored-By: Claude"; then
        echo "❌ ERROR: Git commit contains Claude Code attribution" >&2
        echo "" >&2
        echo "Please remove the following from your commit message:" >&2
        echo "  - 🤖 Generated with [Claude Code](https://claude.com/claude-code)" >&2
        echo "  - Co-Authored-By: Claude <noreply@anthropic.com>" >&2
        echo "" >&2
        echo "Commit messages should contain only technical content." >&2
        exit 1
    fi
fi

# Allow the command to proceed
exit 0
