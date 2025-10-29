#!/bin/bash
# Hook to prevent Claude Code attribution in git commits
# This script checks git commit commands for the Claude attribution text

# Get the full command from parameters
COMMAND="$1"

# Check if this is a git commit command
if [[ "$COMMAND" == *"git commit"* ]]; then
    # Check for Claude attribution patterns
    if echo "$COMMAND" | grep -q "Generated with \[Claude Code\]" || \
       echo "$COMMAND" | grep -q "Co-Authored-By: Claude"; then
        echo "❌ ERROR: Git commit contains Claude Code attribution"
        echo ""
        echo "Please remove the following from your commit message:"
        echo "  - 🤖 Generated with [Claude Code](https://claude.com/claude-code)"
        echo "  - Co-Authored-By: Claude <noreply@anthropic.com>"
        echo ""
        echo "Commit messages should contain only technical content."
        exit 1
    fi
fi

# Allow the command to proceed
exit 0
