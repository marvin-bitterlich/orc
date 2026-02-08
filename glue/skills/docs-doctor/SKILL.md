---
name: docs-doctor
description: Validate documentation against code reality. Use before merging to master, when docs may have drifted, or after significant code changes. Checks DOCS.md index, internal links, CLI commands, skills, and diagram accuracy.
---

# Docs Doctor

Validate documentation against code reality using parallel subagent checks.

## Usage

```
/docs-doctor           (run all checks)
/docs-doctor --fix     (auto-fix simple issues)
```

## When to Use

- Before merging to master (recommended in CLAUDE.md)
- After adding/removing documentation files
- After changing CLI commands or flags
- After modifying database schema (ER diagram)
- After changing shipment guards (lifecycle diagram)

## Check Categories

### 1. Structural Checks
- DOCS.md index matches actual files in docs/
- No orphan files (files not in index)
- No missing files (index references non-existent files)
- Internal markdown links are valid

### 2. Lane Checks
- README.md contains no agent instructions (CLAUDE.md's job)
- CLAUDE.md contains no human onboarding content (README's job)
- Each doc stays in its designated purpose

### 3. Behavioral Checks (Shallow)
- Documented CLI commands exist (`orc <cmd> --help` succeeds)
- Documented flags are valid (parse --help output)
- Referenced skills exist in glue/skills/

### 4. Diagram Checks
- ER diagram in architecture.md matches internal/db/schema.sql
- Lifecycle diagram in common-workflows.md matches guards in internal/core/shipment/guards.go

## Architecture

This skill uses a **fan-out pattern** with haiku subagents for parallel validation:

```
Main Agent (opus)
    ├── Spawn: Structural Check Agent (haiku)
    ├── Spawn: Lane Check Agent (haiku)
    ├── Spawn: CLI Validation Agent (haiku)
    ├── Spawn: ER Diagram Agent (haiku)
    └── Spawn: Lifecycle Diagram Agent (haiku)
         ↓
    Collect findings
         ↓
    Synthesize report
         ↓
    Auto-fix OR escalate
```

## Flow

### Step 1: Spawn Structural Check Agent

Use the Task tool with `subagent_type: "Explore"` and `model: "haiku"`:

```
Prompt: "Check if DOCS.md index matches reality.

1. Read DOCS.md and extract all file paths from the index tables
2. Use Glob to find all .md files in docs/
3. Compare: find missing files (in index but don't exist) and orphan files (exist but not in index)
4. Check each internal link in docs/*.md files resolves

Return findings as:
- missing_files: [list]
- orphan_files: [list]
- broken_links: [list]
- status: 'pass' or 'fail'"
```

### Step 2: Spawn Lane Check Agent

```
Prompt: "Check documentation stays in its designated lane.

1. Read README.md - should NOT contain:
   - References to CLAUDE.md behavior
   - Agent-specific instructions
   - 'Run orc prime' type commands

2. Read CLAUDE.md - should NOT contain:
   - Marketing language
   - Human onboarding ('Welcome to ORC!')
   - Installation instructions for humans

Return findings as:
- readme_violations: [list of lines/content]
- claudemd_violations: [list of lines/content]
- status: 'pass' or 'fail'"
```

### Step 3: Spawn CLI Validation Agent

```
Prompt: "Validate CLI commands referenced in documentation.

1. Read docs/*.md and extract all 'orc <command>' references
2. For each unique command, run 'orc <command> --help' to verify it exists
3. Check documented flags match --help output

Return findings as:
- invalid_commands: [list]
- invalid_flags: [{command, flag}]
- status: 'pass' or 'fail'"
```

### Step 4: Spawn ER Diagram Agent

```
Prompt: "Validate ER diagram matches database schema.

1. Read internal/db/schema.sql
2. Read docs/architecture.md, find the erDiagram mermaid block
3. Compare: are all core tables represented? Are relationships accurate?

Note: Simplified diagram is OK - just verify core entities match.

Return findings as:
- missing_tables: [list]
- incorrect_relationships: [list]
- status: 'pass' or 'fail'"
```

### Step 5: Spawn Lifecycle Diagram Agent

```
Prompt: "Validate shipment lifecycle diagram matches guards.

1. Read internal/core/shipment/guards.go
2. Read docs/common-workflows.md, find the stateDiagram mermaid block
3. Compare: do transitions in diagram match guard logic?

Return findings as:
- missing_transitions: [list]
- invalid_transitions: [list]
- status: 'pass' or 'fail'"
```

### Step 6: Synthesize Findings

Collect all agent results and categorize:

**Auto-fixable:**
- Missing file in DOCS.md index (add it)
- Orphan file not in index (add it)
- Simple broken links (update path)

**Requires Judgment:**
- Lane violations (content in wrong file)
- Diagram mismatches (which is correct?)
- Invalid commands (docs wrong or code wrong?)

### Step 7: Report and Act

**If all pass:**
```
✅ Docs Doctor: All checks passed

Structural: ✓
Lanes: ✓
CLI: ✓
ER Diagram: ✓
Lifecycle: ✓
```

**If issues found:**
```
⚠️ Docs Doctor: Issues found

Structural:
  - Missing from index: docs/new-file.md
  - Broken link: common-workflows.md:45 → nonexistent.md

Lanes:
  - README.md:23 contains agent instruction

CLI:
  - docs/troubleshooting.md references 'orc foobar' (doesn't exist)

Auto-fixing: [list what will be fixed]
Escalating: [list what needs human decision]
```

**With --fix flag:**
- Auto-fix simple issues (update DOCS.md index, fix paths)
- For judgment calls, run `/orc-interview` to get human decision

## Enforcement

This is a **soft gate**:
- Runs on request (not automatic)
- Issues warnings but doesn't block commits
- Documented in CLAUDE.md as pre-merge check
- Git post-merge hook can remind to run

## Notes

- Haiku subagents are cheap and fast for parallel validation
- Each agent has narrow focus, returns structured output
- Main agent orchestrates and synthesizes
- Simple fixes auto-applied, complex decisions escalated
