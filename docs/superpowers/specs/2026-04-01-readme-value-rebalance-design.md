# Design Spec: README Value Rebalance & AI-Agent Story

## Problem

ckeletin-go's most differentiating capability — AI-agent-ready architecture with enforcement by automation — is buried in 10 lines out of 432 in the README. The messaging leads with Reassurance (55%) when it should lead with Smart (currently 40%). Joy is at 3%, nearly absent. The updatable framework story, the project's second differentiator, is present but not connected to the AI story.

Based on a Premium Value Analysis conducted using the Consumer Premium Value Framework, 29% of the product's value evidence comes from AI-agent infrastructure that is virtually unmessaged.

## Context

The March 28 documentation identity overhaul established "scaffold + framework" terminology and cut the README from 983 to 432 lines. That work resolved identity confusion. This spec is the **next evolution**: surfacing the AI-agent story and rebalancing value messaging.

## Approved Decisions

- **Scope**: Full documentation pass (README, AGENTS.md, CONTRIBUTING.md, CONVENSIONS.md rename)
- **Tone**: "Built for humans and AI agents alike" — inclusive, AI as amplifier not gatekeeper
- **"Boss" line**: Keep but demote — moves to a "Who Is This For?" section, no longer the opening
- **Agent-Ready section**: Primary section in README + AI woven into TL;DR and Key Highlights
- **Approach**: "Narrative Arc" — restructure README around Smart → AI → Framework → Reassurance flow

## Emotional Arc

The README follows this narrative flow:

```
"You're capable"         → Smart-led TL;DR
"AI agents amplify you"  → Agent-Ready Architecture section
"The framework evolves"  → What You Get + Framework Updates
"Guardrails protect all" → Enforced Quality in Highlights + existing sections
```

## Files and Changes

### 1. README.md — Restructure

**New section order:**

```
Banner + badges (keep)

## TL;DR
  - Smart-led opening: "production-ready CLI infrastructure... focus on your feature"
  - Bullet 1: Built for humans and AI agents
  - Bullet 2: Updatable framework (linked to AI story)
  - Bullet 3: Read the code in 5 minutes
  - Bullet 4: Ship with ≥85% test coverage / every rule machine-checkable
  - Bullet 5: One command setup
  - Quickstart code block (keep)
  - Bonus: GPL/AGPL blocking (keep)

## What You Get
  - Keep current scaffold + framework explanation verbatim
  - Keep directory tree
  - ADD one paragraph: "AI agents work here too..." connecting framework to AI story

## Key Highlights
  - 8 bullets, rebalanced:
    1. Agent-Ready Architecture (NEW — was absent)
    2. Updatable Framework (linked to AI)
    3. Readable Code
    4. Enforced Quality
    5. Enterprise License Compliance
    6. Task-Based Workflow (now mentions AI agents use same interface)
    7. Reproducible Builds
    8. Crafted to Learn From (Joy touch — replaces "Learn While You Build")

## Agent-Ready Architecture (NEW PRIMARY SECTION)
  - Opening: problem (agents drift) + insight (enforcement by automation)
  - The AI Configuration Stack (visual diagram):
      AGENTS.md → CLAUDE.md → .claude/rules/ → .claude/hooks.json → task check
  - One paragraph per layer explaining its role
  - "Why This Matters" subsection:
      - Determinism (task commands eliminate flag guesswork)
      - Architectural memory (ADRs explain why)
      - Automated enforcement (14 ADRs, machine-checkable)
      - Framework evolution (update improves AI config too)
  - "Using With AI Agents" subsection:
      - Claude Code: automatic, no config needed
      - Cursor / Copilot / Codex: point at AGENTS.md
      - "The pattern is reusable" closing

## Quick Start (keep as-is)

## Architecture
  - Keep existing bullet list
  - Reword the ADR-014 bullet to include AI angle:
    "Enforcement by automation — Every ADR has machine-checkable validation,
    catching violations from humans and AI agents alike"

## Features (keep as-is)

## Getting Started (keep as-is)

## Configuration (keep as-is)

## Commands (keep as-is)

## Development Workflow (keep as-is)

## Framework Updates (keep as-is)

## Who Is This For? (NEW — replaces old Customization section)
  - Paragraph 1: "Your boss needs a CLI tool..." — the relocated Reassurance hook
  - Paragraph 2: Senior dev tired of rebuilding scaffolding
  - Paragraph 3: AI coding agent users who need correct output
  - Paragraph 4: People who want to make their own codebase agent-ready

## Contributing (keep as-is)

## License (keep as-is)
```

**Sections removed:**
- Old "Customization" section — its contents are redistributed:
  - "Changing the Program Name" → already in Quick Start / Taskfile docs
  - "Adding New Commands" → already in CONTRIBUTING.md
  - Old "AI Integration" 10-line subsection → replaced by the new Agent-Ready Architecture primary section

### 2. AGENTS.md — Intro Reframe

**Changes to the top of file only** (rest stays untouched):

- Add blockquote positioning AGENTS.md as a reusable pattern:
  > "This file is a reference implementation. The pattern — structured project guide, behavioral rules, automated hooks, machine-checkable enforcement — works in any codebase."
- Add "built for humans and AI agents alike" to the identity line
- Add enforcement philosophy paragraph: "Every architectural rule in this project is machine-checkable. `task check` is the single gateway..."

### 3. CONTRIBUTING.md — AI Agent Compatibility Subsection

Add a ~6-line "AI Agent Compatibility" subsection after the "Framework vs Project Code" table (around line 76):

```markdown
### AI Agent Compatibility

The framework includes AI agent configuration (AGENTS.md, CLAUDE.md,
`.claude/rules/`, `.claude/hooks.json`) that enables AI coding agents to work
within the project's enforced patterns. When contributing, be aware that changes
to architectural patterns, task commands, or configuration conventions may need
corresponding updates to AGENTS.md and CLAUDE.md so that AI agents stay aligned.
```

### 4. CONVENSIONS.md → CONVENTIONS.md

Rename file to fix the typo. No content changes.

## Content Audit

Content from the current README that moves or is removed:

| Content | Current Location | Destination | Rationale |
|---------|-----------------|-------------|-----------|
| "Your boss needs a CLI tool..." opening | TL;DR lead | Who Is This For? paragraph 1 | Demoted from lead, kept as audience segment |
| "AI Integration" 10-line subsection | Customization | Agent-Ready Architecture (expanded) | Promoted from footnote to primary section |
| "Changing the Program Name" | Customization | Removed (covered by Quick Start) | Redundant |
| "Adding New Commands" snippet | Customization | Removed (covered by CONTRIBUTING.md) | Already exists in contributor guide |

**No content is deleted without a home.** Everything either stays, moves within README, or already exists in CONTRIBUTING.md.

## What NOT to Change

- **CLAUDE.md** — Points to AGENTS.md, no direct identity claims. No changes needed.
- **.claude/hooks.json, .claude/rules/, .claude/settings.local.json** — Working infrastructure, not documentation. Don't touch.
- **ADRs** — Already updated in March overhaul. No changes needed.
- **ARCHITECTURE.md** — Already consistent. No changes needed.
- **CHANGELOG.md** — Historical record. Never retroactively change.
- **Getting Started, Configuration, Commands, Development Workflow, Framework Updates** — Reference material that's already solid.

## Value Messaging Targets

Based on the Premium Value Analysis fingerprint (Smart + Joy + Reassurance):

| Value | Current Share | Target Share | How |
|-------|-------------|-------------|-----|
| Smart | ~40% | ~45% | Lead TL;DR, AI amplification narrative |
| Reassurance | ~55% | ~35% | Still present, no longer leads |
| Joy | ~3% | ~10% | "Crafted to Learn From" highlight, "reasoned" language |
| AI-agent (amplifier) | ~2% | ~15% | New primary section + woven throughout |

## Verification

1. **Content completeness**: Every piece of content in the current README is accounted for in the audit table — kept, moved, or confirmed redundant
2. **`task check`**: Must pass after all changes (no code changes, but verify nothing breaks)
3. **Terminology consistency**: Grep for "skeleton", "template generator", "boilerplate" — should find zero in identity contexts
4. **AI story visibility**: The AI-agent story should appear in TL;DR, What You Get, Key Highlights, and Agent-Ready Architecture — four touchpoints minimum
5. **Framework story preserved**: Framework updatability appears in TL;DR, What You Get, Key Highlights, Agent-Ready Architecture, and Framework Updates — five touchpoints minimum
6. **Line count target**: README should stay in the 430-480 line range (we're removing ~30 lines of Customization and adding ~60 lines of Agent-Ready Architecture + Who Is This For)
7. **Manual read-through**: Read the new README as a first-time visitor. Does the narrative arc flow: capable → amplified → evolving → protected?

## Risks

- **Tone shift may alienate existing audience** — The "boss" line attracted first-time CLI developers. Moving it lower could reduce that audience's engagement. Mitigation: the line is preserved and the new TL;DR still speaks to capability, which serves the same audience without limiting to it.
- **Agent-Ready Architecture section could date quickly** — AI tooling evolves fast. Mitigation: the section focuses on the pattern (enforcement by automation) not specific tool versions.
- **Over-rotation to AI** — Risk of making the project feel like "an AI thing" rather than a Go scaffold. Mitigation: the inclusive "humans and AI agents alike" framing, and the AI section is one of many, not the whole README.

## Implementation Order

1. Rename CONVENSIONS.md → CONVENTIONS.md
2. Update AGENTS.md intro
3. Update CONTRIBUTING.md (add AI Agent Compatibility subsection)
4. Restructure README.md (the main event)
5. Run `task check`
6. Verify with grep and manual read-through
