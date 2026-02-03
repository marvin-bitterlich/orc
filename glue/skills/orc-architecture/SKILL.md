---
name: orc-architecture
description: Maintain ARCHITECTURE.md with C2/C3 structure. Use when you want to document or update the codebase architecture map that ship-plan references.
---

# ORC Architecture

Maintain the ARCHITECTURE.md file that documents the codebase structure using C4 model terminology. This file is the source of truth for ship-plan's C2/C3 mapping.

## Usage

```
/orc-architecture              (update ARCHITECTURE.md)
/orc-architecture --bootstrap  (create initial ARCHITECTURE.md)
```

## When to Use

- When ARCHITECTURE.md doesn't exist yet (bootstrap)
- After significant structural changes to the codebase
- When ship-plan can't find containers/components it expects
- Periodic maintenance to keep architecture map current

## C4 Model Terminology

| Level | What | Examples |
|-------|------|----------|
| C1: System Context | External systems/actors | Users, external APIs |
| C2: Container | Deployable units | CLI binary, database, skill files |
| C3: Component | Modules within containers | cmd/focus.go, internal/ledger/ |
| C4: Code | Implementation details | Functions, classes (not documented here) |

This skill maintains C2 and C3 levels. C4 is too granular for architecture docs.

## Flow

### Step 1: Check Current State

```bash
# Check if ARCHITECTURE.md exists
cat ARCHITECTURE.md 2>/dev/null || echo "No ARCHITECTURE.md found"
```

If file doesn't exist:
```
No ARCHITECTURE.md found. Running bootstrap mode...
```

### Step 2: Explore Codebase

Gather structural information:
```bash
# Top-level directories
ls -la

# Key directories structure
ls cmd/
ls internal/
ls glue/skills/

# Config files
ls *.md *.json 2>/dev/null
```

Build mental map of:
- What containers exist (deployable/distinct units)
- What components are in each container
- How they relate to each other

### Step 3: Compare with Current ARCHITECTURE.md

If ARCHITECTURE.md exists, identify drift:
- New directories/components not documented
- Documented items that no longer exist
- Structural changes (moves, renames)

Present findings:
```
## Architecture Drift Detected

Current ARCHITECTURE.md has:
- Skills container with 12 components
- CLI container with cmd/, internal/

Codebase shows:
- Skills container now has 15 components (+3 new)
- New orc-interview, ship-synthesize, orc-architecture skills

Changes needed:
1. Add orc-interview to Skills components
2. Add ship-synthesize to Skills components
3. Add orc-architecture to Skills components
```

### Step 4: Interview for Confirmation

Use orc-interview format to confirm changes:

```
[Question 1/3]

Found 3 new skills that aren't documented in ARCHITECTURE.md:
- orc-interview
- ship-synthesize
- orc-architecture

These should be added to the Skills container's component list.

1. Approve - add all three
2. Review each individually
3. Skip - don't update architecture
4. Discuss
```

### Step 5: Update ARCHITECTURE.md

After confirmation, update the file:
```bash
# Read current content
# Apply changes
# Write updated content
```

Output:
```
ARCHITECTURE.md updated:
  ✓ Added 3 components to Skills container
  ✓ Removed 2 deprecated components

View changes: cat ARCHITECTURE.md
```

## Bootstrap Mode

When ARCHITECTURE.md doesn't exist, create it:

### Step 1: Deep Codebase Exploration

```bash
# Explore all major directories
find . -type d -maxdepth 2 | grep -v ".git"

# Check for patterns
ls -la cmd/ internal/ glue/ docs/ 2>/dev/null
```

### Step 2: Propose Initial Structure

Present proposed architecture:
```
## Proposed ARCHITECTURE.md

Based on codebase exploration:

### C2 Containers

| Container | Location | Description |
|-----------|----------|-------------|
| CLI | cmd/, internal/ | ORC command-line tool |
| Database | internal/db/ | SQLite ledger storage |
| Skills | glue/skills/ | Claude Code skill definitions |
| Config | .orc/ | Runtime configuration |
| Documentation | *.md | Project documentation |

### C3 Components (Skills container)

| Component | Description |
|-----------|-------------|
| ship-new/ | Create new shipments |
| ship-plan/ | Engineering review and task creation |
| ship-synthesize/ | Knowledge compaction |
| ... | |

Create ARCHITECTURE.md with this structure?
[y]es / [e]dit / [n]o
```

### Step 3: Create File

After confirmation:
```bash
# Write ARCHITECTURE.md with proposed content
```

## ARCHITECTURE.md Template

```markdown
# Architecture

ORC codebase structure documented using C4 model (C2 containers, C3 components).

Last updated: [date]

## C2 Containers

| Container | Location | Description |
|-----------|----------|-------------|
| CLI | cmd/, internal/ | ORC command-line tool |
| Database | internal/db/, schema/ | SQLite ledger with Atlas migrations |
| Skills | glue/skills/ | Claude Code skill definitions |
| Hooks | glue/hooks/ | Git and Claude Code hooks |
| Config | .orc/ | Runtime configuration per workspace |
| Documentation | *.md, docs/ | Project and workflow documentation |

## C3 Components

### CLI (cmd/, internal/)

| Component | Location | Description |
|-----------|----------|-------------|
| Commands | cmd/ | CLI command implementations |
| Ledger | internal/ledger/ | Core domain logic |
| Ports | internal/ports/ | Interface definitions |
| Adapters | internal/adapters/ | Implementation adapters |

### Skills (glue/skills/)

| Component | Description |
|-----------|-------------|
| ship-new | Create new shipments |
| ship-plan | C2/C3 engineering review, task creation |
| ship-synthesize | Knowledge compaction |
| orc-interview | Reusable interview primitive |
| orc-architecture | Maintain this file |
| imp-* | IMP workflow skills |
| ... | |

### Database (internal/db/, schema/)

| Component | Description |
|-----------|-------------|
| schema.sql | Database schema definition |
| migrations/ | Atlas migration files |

## Dependencies

- CLI depends on Database (reads/writes ledger)
- Skills depend on CLI (invoke orc commands)
- Hooks depend on CLI (invoke orc commands)

## Change Log

| Date | Change |
|------|--------|
| YYYY-MM-DD | Initial architecture documentation |
```

## Example Session

```
> /orc-architecture

[checks for ARCHITECTURE.md - not found]

No ARCHITECTURE.md found. Running bootstrap mode...

[explores codebase]
[builds container/component map]

## Proposed ARCHITECTURE.md

Based on codebase exploration:

### C2 Containers
- CLI (cmd/, internal/)
- Database (internal/db/, schema/)
- Skills (glue/skills/)
- Hooks (glue/hooks/)
- Config (.orc/)
- Documentation (*.md, docs/)

### C3 Components (Skills)
- ship-new, ship-plan, ship-synthesize...
- orc-interview, orc-architecture...
- imp-start, imp-plan-create...

Create ARCHITECTURE.md with this structure? [y/e/n]
> y

✓ ARCHITECTURE.md created

This file is now the reference for ship-plan's C2/C3 mapping.
```
