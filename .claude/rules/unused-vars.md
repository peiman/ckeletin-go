# Unused Variables: Investigate Before Deleting

When linting reveals unused variables, DO NOT just delete them.

1. **Investigate first** — Is this variable meant to be used somewhere not yet implemented?
2. **Check context** — Look at surrounding code, function signatures, and commit history for clues
3. **Flag to user** — If it looks like missing functionality, ask before removing
4. **Only then remove** — After confirming it's truly dead code, not an incomplete implementation

Unused variables often signal forgotten implementations, not just dead code.
This has happened before in this codebase — always check before removing.
