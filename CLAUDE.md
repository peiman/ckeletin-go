@AGENTS.md

# Claude Code

Project rules live in [AGENTS.md](AGENTS.md). This file is only the Claude Code behavior that does not belong in the universal guide.

- Change files with the editor. Do not use `sed`, `awk`, or shell redirects to edit code.
- Do not pass `--no-verify` unless the user has approved it after seeing why the hook itself is broken.
- When lint flags an unused variable, read [.claude/rules/unused-vars.md](.claude/rules/unused-vars.md) before deleting it.
- Session setup and tool install: [.claude/rules/claude-setup.md](.claude/rules/claude-setup.md). If install fails: `bash .ckeletin/scripts/install_tools.sh`.
