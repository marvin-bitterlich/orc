---
name: docs-doctor
description: Validate ORC documentation against code reality. Use before merging to master, when docs may have drifted, or after significant code changes. Checks internal links, CLI commands, skills, and diagram accuracy.
---

# Docs Doctor

Validate ORC documentation against code reality using parallel subagent checks.

**Note:** This is an ORC-specific skill. It validates ORC's documentation structure and should only be used in the ORC repository.

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
- Internal markdown links in docs/*.md are valid
- No broken cross-references between documentation files

### 2. Lane Checks
- README.md contains no agent instructions (CLAUDE.md's job)
- CLAUDE.md contains no human onboarding content (README's job)
- Each doc stays in its designated purpose

### 3. Behavioral Checks (Shallow)
- Documented CLI commands exist (`orc <cmd> --help` succeeds)
- Documented flags are valid (parse --help output)
- Referenced skills exist in .claude/skills/ or glue/skills/

### 4. Diagram Checks
- ER diagram in docs/architecture.md represents core tables from internal/db/schema.sql
- Lifecycle diagram in docs/shipment-lifecycle.md represents valid subset of states from internal/core/shipment/guards.go

**Note:** Diagrams are intentionally simplified. Validation checks that diagram states/tables are a subset of code reality, not an exact match.

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
Prompt: "Check internal documentation links.

1. Use Glob to find all .md files in docs/
2. For each file, extract internal markdown links (links to other docs/*.md files)
3. Verify each linked file exists

Return findings as:
- broken_links: [{file, line, target}]
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
Prompt: "Validate ER diagram represents database schema.

1. Read internal/db/schema.sql
2. Read docs/architecture.md, find the erDiagram mermaid block
3. Check that tables shown in diagram exist in schema (subset validation)
4. Check that relationships shown are accurate

Note: Diagram is intentionally simplified - not all tables need to be shown.

Return findings as:
- invalid_tables: [tables in diagram but not in schema]
- incorrect_relationships: [list]
- status: 'pass' or 'fail'"
```

### Step 5: Spawn Lifecycle Diagram Agent

```
Prompt: "Validate shipment lifecycle diagram represents valid states.

1. Read internal/core/shipment/guards.go
2. Read docs/shipment-lifecycle.md, find the stateDiagram mermaid block
3. Check that states in diagram are valid states from guards.go (subset validation)
4. Check that transitions shown are valid according to guards

Note: Diagram is intentionally simplified - not all states need to be shown.

Return findings as:
- invalid_states: [states in diagram but not in guards]
- invalid_transitions: [transitions that guards don't allow]
- status: 'pass' or 'fail'"
```

### Step 6: Synthesize Findings

Collect all agent results and categorize:

**Auto-fixable:**
- Simple broken links (update path)

**Requires Judgment:**
- Lane violations (content in wrong file)
- Diagram mismatches (which is correct?)
- Invalid commands (docs wrong or code wrong?)

### Step 7: Report and Act

**If all pass:**
```
Docs Doctor: All checks passed

Structural: pass
Lanes: pass
CLI: pass
ER Diagram: pass
Lifecycle: pass
```

**If issues found:**
```
Docs Doctor: Issues found

Structural:
  - Broken link: common-workflows.md:45 -> nonexistent.md

Lanes:
  - README.md:23 contains agent instruction

CLI:
  - docs/troubleshooting.md references 'orc foobar' (doesn't exist)

Auto-fixing: [list what will be fixed]
Escalating: [list what needs human decision]
```

**With --fix flag:**
- Auto-fix simple issues (fix paths)
- For judgment calls, run `/orc-interview` to get human decision

## Enforcement

This is a **soft gate**:
- Runs on request (not automatic)
- Issues warnings but doesn't block commits
- Documented in CLAUDE.md as pre-merge check

## Notes

- Haiku subagents are cheap and fast for parallel validation
- Each agent has narrow focus, returns structured output
- Main agent orchestrates and synthesizes
- Simple fixes auto-applied, complex decisions escalated
- Diagrams use subset validation, not exact matching
