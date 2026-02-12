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
- After changing the Makefile bootstrap target, `orc bootstrap` command, or orc-first-run skill

## Check Categories

### 1. Structural Checks
- Internal markdown links in docs/**/*.md are valid
- No broken cross-references between documentation files
- `docs/guide/glossary.md` exists (not `glossary/` directory)
- Glossary contains term definitions (no mermaid diagrams)

### 2. Lane Checks
- README.md contains no agent instructions (CLAUDE.md's job)
- CLAUDE.md contains no human onboarding content (README's job)
- Each doc stays in its designated purpose

### 3. Behavioral Checks (Shallow)
- CLI commands referenced anywhere (docs, skills, Makefile) exist (`orc <cmd> --help` succeeds)
- CLI flags used anywhere match actual `--help` output for that command
- Referenced skills exist in .claude/skills/ or glue/skills/

### 4. Schema Checks
- Required directories exist: `docs/dev/`, `docs/guide/`, `docs/reference/`
- Required files present per `docs/reference/docs.md` schema
- No top-level `.md` files directly in `docs/` (all must be in subdirectories)
- No unexpected subdirectories under `docs/`
- No orphan files outside the defined schema

### 5. Getting Started Coherence Check
- `docs/guide/getting-started.md` describes a phased setup flow backed by real code
- Phase 2 claims (what `make bootstrap` does) match the Makefile `bootstrap` target
- Phase 3 claims (what `orc bootstrap` does) match `internal/cli/bootstrap_cmd.go` and `glue/skills/orc-first-run/SKILL.md`
- CLI flags used in examples and skills match actual `--help` output (e.g., `--path` not `--local-path`)
- The guide's description of what each phase creates/leaves behind matches code reality

### 6. Diagram Checks
- ER diagram in docs/reference/architecture.md represents core tables from internal/db/schema.sql
- Lifecycle diagram in docs/reference/shipment-lifecycle.md represents valid subset of states from internal/core/shipment/guards.go

**Note:** Diagrams are intentionally simplified. Validation checks that diagram states/tables are a subset of code reality, not an exact match.

## Architecture

This skill uses a **fan-out pattern** with haiku subagents for parallel validation:

```
Main Agent (opus)
    ├── Spawn: Structural Check Agent (haiku)
    ├── Spawn: Lane Check Agent (haiku)
    ├── Spawn: CLI Validation Agent (haiku)
    ├── Spawn: Schema Validation Agent (haiku)
    ├── Spawn: Getting Started Coherence Agent (haiku)
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
2. For each file, extract internal markdown links (links to other docs/**/*.md files)
3. Verify each linked file exists

Return findings as:
- broken_links: [{file, line, target}]
- status: 'pass' or 'fail'"
```

### Step 1b: Spawn Glossary Check Agent

```
Prompt: "Validate glossary structure.

1. Check that docs/guide/glossary.md exists
2. Check that docs/glossary/ directory does NOT exist
3. Read glossary.md and verify:
   - Contains term definitions (look for **Term** pattern)
   - No mermaid code blocks (no complex diagrams)

Return findings as:
- glossary_exists: boolean
- glossary_dir_exists: boolean (should be false)
- has_mermaid: boolean (should be false)
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

1. Scan all sources that reference orc CLI commands:
   - docs/**/*.md
   - glue/skills/**/*.md
   - .claude/skills/**/*.md
   - Makefile (the bootstrap target and any other targets that call orc)
2. Extract all 'orc <command>' references and any --flags used with them
3. For each unique command, run 'orc <command> --help' to verify it exists
4. For each flag, verify it appears in the command's --help output

Return findings as:
- invalid_commands: [{file, line, command}]
- invalid_flags: [{file, line, command, flag_used, valid_flags}]
- status: 'pass' or 'fail'"
```

### Step 4: Spawn Schema Validation Agent

```
Prompt: "Validate docs directory structure against the docs schema.

1. Read docs/reference/docs.md to understand the expected structure
2. Use Glob to find all .md files under docs/
3. Check:
   a. Required directories exist: docs/dev/, docs/guide/, docs/reference/
   b. No .md files directly in docs/ (all must be in subdirectories)
   c. No unexpected subdirectories under docs/ (only dev/, guide/, reference/)
   d. All files marked 'required: yes' in the schema exist
   e. Every .md file found is listed in the schema (no orphans)

Return findings as:
- missing_dirs: [list of missing required directories]
- top_level_files: [.md files directly in docs/]
- unexpected_dirs: [subdirectories not in schema]
- missing_required: [{dir, file}]
- orphan_files: [{dir, file}]
- status: 'pass' or 'fail'"
```

### Step 5: Spawn Getting Started Coherence Agent

```
Prompt: "Validate that docs/guide/getting-started.md accurately describes the bootstrap flow.

Cross-reference against three source-of-truth files:

1. Read the Makefile 'bootstrap' target (search for '^bootstrap:')
2. Read internal/cli/bootstrap_cmd.go
3. Read glue/skills/orc-first-run/SKILL.md
4. Read docs/guide/getting-started.md

Then check:

a. Phase 2 claims (what 'make bootstrap' does):
   - Every action claimed in the guide has a corresponding line in the Makefile target
   - No significant Makefile actions are missing from the guide
   - Artifacts claimed (e.g. FACT-001, REPO-001, directories) match what the Makefile creates
   - CLI flags in examples match actual flag names (check against 'orc repo create --help', etc.)

b. Phase 3 claims (what 'orc bootstrap' does):
   - Guide's description matches what bootstrap_cmd.go actually does (launches Claude with directive)
   - Claims about what the first-run skill creates match glue/skills/orc-first-run/SKILL.md
   - Entity types mentioned (commission, workshop, workbench, shipment) match the skill's flow

c. Phase ordering:
   - Guide presents phases in correct dependency order
   - Each phase's 'what you have after' description is consistent with the next phase's starting assumptions

d. Cross-file flag consistency:
   - CLI flags used in getting-started.md match flags used in orc-first-run SKILL.md
   - Both match actual --help output for those commands

Return findings as:
- phase2_mismatches: [{claim, reality}]
- phase3_mismatches: [{claim, reality}]
- ordering_issues: [list]
- flag_mismatches: [{file, flag_used, correct_flag}]
- status: 'pass' or 'fail'"
```

### Step 6: Spawn ER Diagram Agent

```
Prompt: "Validate ER diagram represents database schema.

1. Read internal/db/schema.sql
2. Read docs/reference/architecture.md, find the erDiagram mermaid block
3. Check that tables shown in diagram exist in schema (subset validation)
4. Check that relationships shown are accurate

Note: Diagram is intentionally simplified - not all tables need to be shown.

Return findings as:
- invalid_tables: [tables in diagram but not in schema]
- incorrect_relationships: [list]
- status: 'pass' or 'fail'"
```

### Step 7: Spawn Lifecycle Diagram Agent

```
Prompt: "Validate shipment lifecycle diagram represents valid states.

1. Read internal/core/shipment/guards.go
2. Read docs/reference/shipment-lifecycle.md, find the stateDiagram mermaid block
3. Check that states in diagram are valid states from guards.go (subset validation)
4. Check that transitions shown are valid according to guards

Note: Diagram is intentionally simplified - not all states need to be shown.

Return findings as:
- invalid_states: [states in diagram but not in guards]
- invalid_transitions: [transitions that guards don't allow]
- status: 'pass' or 'fail'"
```

### Step 8: Synthesize Findings

Collect all agent results and categorize:

**Auto-fixable:**
- Simple broken links (update path)

**Requires Judgment:**
- Lane violations (content in wrong file)
- Diagram mismatches (which is correct?)
- Invalid commands (docs wrong or code wrong?)

### Step 9: Report and Act

**If all pass:**
```
Docs Doctor: All checks passed

Structural: pass
Glossary: pass
Lanes: pass
CLI: pass
Schema: pass
Getting Started: pass
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
  - docs/guide/troubleshooting.md references 'orc foobar' (doesn't exist)

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
