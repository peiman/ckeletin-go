# README Value Rebalance & AI-Agent Story Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restructure project documentation to lead with Smart + AI-agent readiness + updatable framework, rebalancing away from Reassurance-heavy messaging.

**Architecture:** Four files change: CONVENSIONS.md rename, AGENTS.md intro reframe, CONTRIBUTING.md AI subsection, README.md full restructure. No code changes. The README follows a "Narrative Arc" structure: capable → amplified → evolving → protected.

**Tech Stack:** Markdown documentation only. `task check` for verification. `git mv` for rename.

**Spec:** `docs/superpowers/specs/2026-04-01-readme-value-rebalance-design.md`

---

### Task 1: Rename CONVENSIONS.md → CONVENTIONS.md

**Files:**
- Rename: `CONVENSIONS.md` → `CONVENTIONS.md`

- [ ] **Step 1: Rename the file**

```bash
git mv CONVENSIONS.md CONVENTIONS.md
```

- [ ] **Step 2: Check for references to the old filename**

```bash
grep -rn "CONVENSIONS" *.md docs/ .claude/ .ckeletin/ CLAUDE.md AGENTS.md CONTRIBUTING.md README.md
```

Expected: No hits. If any references exist, update them to `CONVENTIONS.md`.

- [ ] **Step 3: Commit**

```bash
git add CONVENTIONS.md
git commit -m "docs: fix CONVENSIONS.md → CONVENTIONS.md typo"
```

---

### Task 2: Reframe AGENTS.md Intro

**Files:**
- Modify: `AGENTS.md:1-18`

The rest of AGENTS.md (line 19 onward) stays untouched. Only the top section changes.

- [ ] **Step 1: Replace lines 1-18 of AGENTS.md with the new intro**

Replace the current content:

```markdown
# ckeletin-go — Project Guide for AI Agents

## About This Project

**ckeletin-go** is a production-ready Go CLI scaffold powered by an updatable framework layer.

The `.ckeletin/` directory contains the **framework** — config registry, logging, validation scripts, task definitions, and ADRs (000-099). Your code lives in `cmd/`, `internal/`, `pkg/`. Framework updates via `task ckeletin:update` without touching your code.

Key characteristics:
- Ultra-thin command pattern (commands ≤30 lines, logic in `internal/`)
- Centralized configuration registry with auto-generated constants
- Structured logging with Zerolog (dual console + file output)
- Bubble Tea for interactive UIs
- Dependency injection over mocking
- 85% minimum test coverage, enforced by CI
- Public reusable packages in `pkg/` (e.g., `checkmate` for beautiful CLI output)

**Platform:** macOS and Linux. Windows is not officially supported.
```

With the new content:

```markdown
# ckeletin-go — Project Guide for AI Agents

> **This file is a reference implementation.** The pattern — structured project
> guide, behavioral rules, automated hooks, machine-checkable enforcement — works
> in any codebase. See the README for how the pieces fit together.

## About This Project

**ckeletin-go** is a production-ready Go CLI scaffold powered by an updatable framework layer — built for humans and AI agents alike.

The `.ckeletin/` directory contains the **framework** — config registry, logging, validation scripts, task definitions, and ADRs (000-099). Your code lives in `cmd/`, `internal/`, `pkg/`. Framework updates via `task ckeletin:update` without touching your code.

Every architectural rule in this project is machine-checkable. `task check` is the single gateway — run it before every commit. If it passes, the code is correct regardless of who wrote it.

Key characteristics:
- Ultra-thin command pattern (commands ≤30 lines, logic in `internal/`)
- Centralized configuration registry with auto-generated constants
- Structured logging with Zerolog (dual console + file output)
- Bubble Tea for interactive UIs
- Dependency injection over mocking
- 85% minimum test coverage, enforced by CI
- Public reusable packages in `pkg/` (e.g., `checkmate` for beautiful CLI output)

**Platform:** macOS and Linux. Windows is not officially supported.
```

- [ ] **Step 2: Verify the rest of the file is untouched**

Read AGENTS.md line 20 onward and confirm it matches the original `## Commands` section and everything below it.

- [ ] **Step 3: Commit**

```bash
git add AGENTS.md
git commit -m "docs: reframe AGENTS.md as reusable reference implementation

- Add blockquote positioning AGENTS.md as a reusable pattern
- Add 'built for humans and AI agents alike' to identity line
- Add enforcement philosophy paragraph"
```

---

### Task 3: Add AI Agent Compatibility Subsection to CONTRIBUTING.md

**Files:**
- Modify: `CONTRIBUTING.md:77-80` (insert after the two-tier ADR system paragraph)

- [ ] **Step 1: Insert the new subsection after line 79**

After the current content at line 79 (`- Project ADRs (100+) in `docs/adr/` — your project-specific decisions`), insert:

```markdown

### AI Agent Compatibility

The framework includes AI agent configuration (`AGENTS.md`, `CLAUDE.md`, `.claude/rules/`, `.claude/hooks.json`) that enables AI coding agents to work within the project's enforced patterns. When contributing, be aware that changes to architectural patterns, task commands, or configuration conventions may need corresponding updates to `AGENTS.md` and `CLAUDE.md` so that AI agents stay aligned.
```

This goes between the "Two-tier ADR system" paragraph and the `## Development Workflow` heading.

- [ ] **Step 2: Verify surrounding context is intact**

Read CONTRIBUTING.md lines 62-90 and confirm:
- The "Framework vs Project Code" table is untouched above
- The new subsection appears after the two-tier ADR paragraph
- `## Development Workflow` follows below

- [ ] **Step 3: Commit**

```bash
git add CONTRIBUTING.md
git commit -m "docs: add AI Agent Compatibility subsection to CONTRIBUTING.md

- Alert contributors that AI config exists
- Note that pattern changes may need AGENTS.md/CLAUDE.md updates"
```

---

### Task 4: Restructure README.md — TL;DR and What You Get

This is the main event. The README restructure is split across Tasks 4-7 to keep each task focused. Work top-to-bottom through the file.

**Files:**
- Modify: `README.md:1-76`

- [ ] **Step 1: Replace the TL;DR section (lines 28-49)**

Find the current TL;DR:

```markdown
## TL;DR

**Your boss needs a CLI tool by next sprint. You've never built one.**

ckeletin-go gives you production-ready infrastructure so you can focus on YOUR feature, not learning Cobra.

- **Read the code in 5 minutes** - Ultra-thin commands (~20 lines each). No framework magic to decode.
- **Ship with ≥85% test coverage** - Hundreds of real tests. Integration + unit. You won't break production.
- **One command setup** - `task init name=myapp module=...` updates 40+ files. Start coding in 2 minutes.
- **Keep the framework updated** - `task ckeletin:update` pulls improvements without touching your code.
- **Learn as you build** - ADRs explain every decision. Level up while shipping.

**Quickstart:**
```bash
git clone https://github.com/peiman/ckeletin-go.git && cd ckeletin-go
task setup && task init name=myapp module=github.com/you/myapp
task build && ./myapp ping  # All tests passed
```

You just built a CLI with better architecture than most production codebases.

**Bonus:** Automatic GPL/AGPL blocking prevents license contamination.
```

Replace with:

```markdown
## TL;DR

ckeletin-go gives you production-ready CLI infrastructure — clean architecture, enforced patterns, and an updatable framework — so you can focus on your feature.

- **Built for humans and AI agents** — `AGENTS.md`, `CLAUDE.md`, hooks, and automated enforcement mean AI coding agents produce correct, well-structured code from day one
- **Updatable framework** — `.ckeletin/` updates independently via `task ckeletin:update`. Your code is never touched. AI agent infrastructure improves automatically
- **Read the code in 5 minutes** — Ultra-thin commands (~20 lines each). No framework magic to decode
- **Ship with ≥85% test coverage** — Hundreds of real tests. Integration + unit. Every rule is machine-checkable
- **One command setup** — `task init name=myapp module=...` updates 40+ files. Start coding in 2 minutes

**Quickstart:**
```bash
git clone https://github.com/peiman/ckeletin-go.git && cd ckeletin-go
task setup && task init name=myapp module=github.com/you/myapp
task build && ./myapp ping
```

**Bonus:** Automatic GPL/AGPL blocking prevents license contamination.
```

- [ ] **Step 2: Add the AI paragraph to the What You Get section (after line 75)**

Find the end of the What You Get section. After the line:

```markdown
**The framework** keeps working: enforced architecture, validated patterns, type-safe config, structured logging — all updated independently of your code via `task ckeletin:update`.
```

Insert:

```markdown

**AI agents work here too.** The framework includes layered AI configuration — `AGENTS.md` for any AI assistant, `CLAUDE.md` for Claude Code, automated hooks, and behavioral rules — so coding agents follow the same enforced patterns you do. When the framework updates, your AI agent's effectiveness improves with it.
```

- [ ] **Step 3: Verify TL;DR and What You Get read correctly**

Read README.md lines 28-80. Confirm:
- TL;DR leads with Smart ("focus on your feature")
- AI agents are bullet #1
- Framework updatability is bullet #2, linked to AI
- "Boss" line is gone from TL;DR
- What You Get has the new AI paragraph as the final paragraph
- Directory tree and scaffold/framework explanation are untouched

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: rewrite TL;DR as Smart-led, add AI to What You Get

- Lead with capability, not anxiety
- AI agents as bullet #1 in TL;DR
- Framework updatability linked to AI story
- Add 'AI agents work here too' paragraph to What You Get"
```

---

### Task 5: Restructure README.md — Key Highlights

**Files:**
- Modify: `README.md` — the Key Highlights section (currently lines 79-88)

- [ ] **Step 1: Replace the Key Highlights section**

Find the current section:

```markdown
## Key Highlights

- **Readable Architecture**: Ultra-thin commands (~20 lines each) — understand and modify code in minutes
- **Production-Ready Testing**: ≥85% test coverage enforced. Integration + unit tests. CI fails if quality drops
- **One-Command Customization**: `task init` updates 40+ files automatically
- **Updatable Framework**: `.ckeletin/` layer updates independently — your code is never affected
- **Enterprise License Compliance**: Automated GPL/AGPL blocking prevents legal contamination
- **Reproducible Builds**: Pinned tool versions ensure identical results across dev/CI
- **Task-Based Workflow**: Single source of truth for all commands ([ADR-000](.ckeletin/docs/adr/000-task-based-single-source-of-truth.md))
- **Learn While You Build**: 14 ADRs explain every architectural decision
```

Replace with:

```markdown
## Key Highlights

- **Agent-Ready Architecture**: Layered AI configuration (`AGENTS.md` → `CLAUDE.md` → hooks → enforcement) means coding agents produce correct code within your architecture — not despite it
- **Updatable Framework**: `.ckeletin/` updates independently of your code. Patterns, tooling, and AI agent infrastructure evolve together
- **Readable Code**: Ultra-thin commands (~20 lines each) — understand and modify in minutes
- **Enforced Quality**: ≥85% test coverage, automated architecture validation, pre-commit hooks. Every rule is machine-checkable
- **Enterprise License Compliance**: Automated GPL/AGPL blocking prevents legal contamination
- **Task-Based Workflow**: Single source of truth for all commands — local, CI, and AI agents use the same interface ([ADR-000](.ckeletin/docs/adr/000-task-based-single-source-of-truth.md))
- **Reproducible Builds**: Pinned tool versions ensure identical results everywhere
- **Crafted to Learn From**: 14 ADRs explain every architectural decision. The codebase isn't just functional — it's reasoned
```

- [ ] **Step 2: Verify the highlights read correctly**

Read the section. Confirm:
- Agent-Ready Architecture is bullet #1
- Updatable Framework is #2, linked to AI
- Task-Based Workflow mentions AI agents
- Last bullet has Joy touch ("reasoned")
- "One-Command Customization" is gone (covered in TL;DR)
- ADR-000 link is preserved

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: rebalance Key Highlights — AI and framework lead

- Agent-Ready Architecture as bullet #1
- Updatable Framework linked to AI story
- Task-Based Workflow mentions AI agents
- Add Joy touch to 'Crafted to Learn From'
- Remove 'One-Command Customization' (covered in TL;DR)"
```

---

### Task 6: Restructure README.md — Add Agent-Ready Architecture Section

**Files:**
- Modify: `README.md` — insert new section after Key Highlights, before Quick Start

- [ ] **Step 1: Insert the Agent-Ready Architecture section**

After the closing `---` of the Key Highlights section and before `## Quick Start`, insert the following new section:

```markdown
## Agent-Ready Architecture

Most scaffolds produce code that AI agents can write *in* but not write *well in*. Agents guess at conventions, misconfigure flags, and drift from intended patterns. ckeletin-go solves this with **enforcement by automation** — every architectural rule is machine-checkable, so violations are caught whether the code comes from a human or an AI.

### The AI Configuration Stack

```
AGENTS.md          → Universal project guide (any AI assistant)
CLAUDE.md          → Claude Code-specific behavioral rules
.claude/rules/     → Granular rules loaded automatically
.claude/hooks.json → Auto-installs tools, validates commits
task check         → Single gateway that catches all violations
```

**`AGENTS.md`** gives any AI agent complete project context: architecture, commands, conventions, testing thresholds, and decision trees. It's structured as a specification, not prose — designed for machine consumption.

**`CLAUDE.md`** adds Claude Code-specific rules: mandatory task commands, code placement decision trees, priority cascade (Security → License → Correctness → Coverage → Style).

**Hooks and enforcement** close the loop. SessionStart hooks auto-install tools. Pre-commit hooks validate changes. `task check` runs the same quality gates regardless of who wrote the code.

### Why This Matters

- **Determinism**: `task test` always runs the right flags. Agents don't guess `go test -race -coverprofile=... -count=1 ./...`
- **Architectural memory**: ADRs explain *why* patterns exist, preventing agents from optimizing away guardrails they don't understand
- **Automated enforcement**: 14 ADRs, each with machine-checkable validation. No honor system
- **Framework evolution**: `task ckeletin:update` improves the AI configuration alongside everything else

### Using With AI Agents

**Claude Code**: Reads `CLAUDE.md` and `.claude/rules/` automatically. Hooks fire on session start. No configuration needed.

**Cursor / Copilot / Codex**: Point your agent at `AGENTS.md` for full project context. The task-based workflow and automated enforcement work with any tool.

**The pattern is reusable.** The `AGENTS.md` → rules → hooks → enforcement approach works in any codebase. ckeletin-go is a reference implementation.
```

- [ ] **Step 2: Verify the section reads correctly**

Read the new section. Confirm:
- Opens with problem + insight
- Stack diagram is present and readable
- Each layer gets a one-paragraph explanation
- "Why This Matters" has 4 bullets: determinism, memory, enforcement, evolution
- "Using With AI Agents" covers Claude Code, Cursor/Copilot/Codex, and the reusable pattern
- Quick Start section follows immediately after

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add Agent-Ready Architecture as primary README section

- Explain enforcement-by-automation philosophy
- Show the AI configuration stack diagram
- Cover why determinism, architectural memory, and enforcement matter
- Practical guide for Claude Code, Cursor, Copilot, Codex
- Position the pattern as reusable for any codebase"
```

---

### Task 7: Restructure README.md — Architecture Bullet + Replace Customization with Who Is This For

**Files:**
- Modify: `README.md` — Architecture section + Customization section

- [ ] **Step 1: Update the ADR-014 bullet in the Architecture section**

Find the current bullet:

```markdown
- **Every ADR enforced** — By automation, not code review alone ([ADR-014](.ckeletin/docs/adr/014-adr-enforcement-policy.md))
```

Replace with:

```markdown
- **Enforcement by automation** — Every ADR has machine-checkable validation, catching violations from humans and AI agents alike ([ADR-014](.ckeletin/docs/adr/014-adr-enforcement-policy.md))
```

- [ ] **Step 2: Replace the Customization section with Who Is This For**

Find the current Customization section:

```markdown
## Customization

### Changing the Program Name

In `Taskfile.yml`:
```yaml
vars:
  BINARY_NAME: myapp
```

Then: `task build && ./myapp ping`

### Adding New Commands

```bash
task generate:command name=hello
```

Creates `cmd/hello.go` (ultra-thin wrapper) and `internal/config/commands/hello_config.go` (config metadata). See [CONTRIBUTING.md](CONTRIBUTING.md) for the full step-by-step guide.

### AI Integration

This project includes `AGENTS.md` (universal AI agent guide) and `CLAUDE.md` (Claude Code-specific rules):

- Task-based workflow commands
- Architecture and ADR references
- Code quality standards and coverage requirements
- Testing conventions and git workflow

[Claude Code](https://docs.anthropic.com/en/docs/claude-code) reads `CLAUDE.md` automatically. Other AI assistants (Cursor, Copilot, Codex) can use `AGENTS.md`.
```

Replace with:

```markdown
## Who Is This For?

**Your boss needs a CLI tool by next sprint. You've never built one.**
Clone, `task init`, and you have production-ready infrastructure in 2 minutes. The ADRs teach you the patterns as you build.

**You're a senior dev who's tired of rebuilding the same scaffolding.**
The updatable framework means you set up once and receive improvements over time. The enforced patterns mean your code stays clean even as the team grows.

**You use AI coding agents and need them to produce correct code.**
The layered AI configuration — `AGENTS.md`, `CLAUDE.md`, hooks, enforcement — means agents work within your architecture, not around it. This is what "agent-ready" looks like.

**You want to make your own codebase agent-ready.**
Study the pattern: `AGENTS.md` → behavioral rules → automated hooks → machine-checkable enforcement. It works in any project.
```

- [ ] **Step 3: Verify the changes**

Read the Architecture section and confirm the updated ADR-014 bullet.
Read the new Who Is This For section and confirm:
- "Boss" line is present as paragraph 1
- Four audience segments are distinct
- Old Customization content (program name, adding commands, AI Integration subsection) is gone
- "Changing the Program Name" is already covered in Quick Start / Taskfile docs
- "Adding New Commands" is already in CONTRIBUTING.md
- "AI Integration" is replaced by the full Agent-Ready Architecture section

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: update Architecture bullet, replace Customization with Who Is This For

- Reword ADR-014 bullet to include AI angle
- Replace Customization grab-bag with audience-focused section
- Relocate 'boss' line as one of four audience segments
- Remove redundant content (covered by CONTRIBUTING.md and Agent-Ready Architecture)"
```

---

### Task 8: Verification

**Files:**
- All modified files

- [ ] **Step 1: Run task check**

```bash
task check
```

Expected: All checks pass. No code changes were made, so this should be clean. If it fails, read the error output and fix the root cause.

- [ ] **Step 2: Check for stale terminology**

```bash
grep -rn "skeleton\|template generator\|boilerplate" README.md AGENTS.md CONTRIBUTING.md CONVENTIONS.md
```

Expected: No hits in identity contexts. "boilerplate" may appear in ADR references as generic English usage — that's acceptable.

- [ ] **Step 3: Verify AI story has 4+ touchpoints in README**

```bash
grep -n "AI\|agent" README.md
```

Confirm the AI-agent story appears in:
1. TL;DR (bullet #1)
2. What You Get (closing paragraph)
3. Key Highlights (bullets #1, #2, #6)
4. Agent-Ready Architecture (full section)
5. Who Is This For (paragraphs #3, #4)

- [ ] **Step 4: Verify framework story has 5+ touchpoints in README**

```bash
grep -n "framework\|ckeletin:update\|\.ckeletin" README.md
```

Confirm framework updatability appears in:
1. TL;DR (bullet #2)
2. What You Get (full section)
3. Key Highlights (bullet #2)
4. Agent-Ready Architecture ("Framework evolution" bullet)
5. Framework Updates (existing section)
6. Who Is This For (paragraph #2)

- [ ] **Step 5: Manual read-through**

Read the full README top to bottom as a first-time visitor. Verify:
- The narrative arc flows: capable → amplified → evolving → protected
- No sections feel disconnected or redundant
- The "boss" line in Who Is This For reads naturally as one audience segment
- The Agent-Ready Architecture section is informative without being preachy

- [ ] **Step 6: Check README line count**

```bash
wc -l README.md
```

Expected: 430-480 lines (we removed ~30 lines of Customization and added ~60 lines of Agent-Ready Architecture + Who Is This For).

- [ ] **Step 7: Final commit (if any fixups needed)**

If the verification steps surfaced any issues and fixes were made:

```bash
git add -A
git commit -m "docs: fixups from verification pass"
```
