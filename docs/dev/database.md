# Database & Schema Management

ORC uses SQLite with [Atlas](https://atlasgo.io/) for declarative schema migrations.

## Source of Truth

The single source of truth for the database schema is `internal/db/schema.sql`.

## Why Atlas?

Hand-rolled SQLite migrations repeatedly caused FK reference corruption during table renames. Atlas catches these at validation time -- you can't even *define* a schema with dangling FK references.

### Installation

```bash
brew install ariga/tap/atlas
```

## Core Workflow

**1. Edit the schema:**
Edit `internal/db/schema.sql` (the single source of truth)

**2. Preview changes:**
```bash
make schema-diff
```

**3. Apply changes:**
```bash
make schema-apply
```

**4. Inspect current database:**
```bash
make schema-inspect
```

All Makefile targets use `--env local` from `atlas.hcl`, which handles:
- Schema source (`internal/db/schema.sql`)
- Database URL (`~/.orc/orc.db`)
- SQLite autoindex exclusion (required due to Atlas bug)

### Key Behaviors

- **FK validation**: Atlas refuses to process schemas with dangling FK references
- **Auto-generates SQL**: Handles SQLite's rename-recreate dance automatically
- **Dependency ordering**: Knows which tables reference which, applies changes in safe order
- **Data preservation**: Copies data during table recreates

### Golden Rule

**Never write migration SQL by hand.** Edit `schema.sql`, let Atlas diff and apply.

## Two-Database Model

ORC uses a two-database model to prevent accidental modification of production data.

| Database | Path | Command | Purpose |
|----------|------|---------|---------|
| Production | `~/.orc/orc.db` | `orc` | Real commissions, shipments, tasks |
| Workbench | `.orc/workbench.db` | `orc-dev` | Isolated development per workbench |

**There is no shared dev database.** Each workbench has its own isolated database that must be explicitly created.

### Setup: Create Workbench Database

Before using `orc-dev`, create a workbench-local database:

```bash
make setup-workbench
```

This creates `.orc/workbench.db` with fresh fixtures:
- 3 tags, 2 repos
- 2 factories, 3 workshops, 2 workbenches
- 3 commissions, 5 shipments, 10 tasks
- 2 tomes, 4 notes

### Using orc-dev

The `orc-dev` command **requires** a workbench database:

```bash
orc-dev summary          # Works if .orc/workbench.db exists
orc-dev dev doctor       # Verifies environment
```

If no workbench DB exists, `orc-dev` will error:
```
Error: No workbench DB found at .orc/workbench.db
Run: make setup-workbench
```

### Resetting the Workbench Database

To start fresh:

```bash
make setup-workbench     # Recreates .orc/workbench.db with fixtures
```

## IMP Schema Modification Workflow

When developing ORC itself and modifying the database schema:

**1. Create/reset workbench DB:**
```bash
make setup-workbench
```

**2. Edit the schema:**
```bash
$EDITOR internal/db/schema.sql
```

**3. Preview changes against workbench DB:**
```bash
make schema-diff-workbench
```

**4. Apply to workbench DB:**
```bash
make schema-apply-workbench
```

**5. Test with workbench DB:**
```bash
orc-dev summary          # Manual verification
make test                # Automated tests (use in-memory DB)
```

**6. Verify lint passes:**
```bash
make lint
```

**7. Commit your changes** (pre-commit hook enforces tests + lint)

### Golden Rules

1. **Always `make setup-workbench` before schema work** -- Creates isolated DB
2. **Use `make schema-*-workbench` for iteration** -- Fast feedback loop
3. **Let Atlas generate SQL** -- Never write migration SQL by hand
4. **Tests must pass before commit** -- Pre-commit hook enforces this

## Data & Config Changes

**Don't overload "migration"** -- these are distinct operations:

| Term | When | Where | Runs |
|------|------|-------|------|
| **Schema change** | Development | `internal/db/schema.sql` + Atlas | Per-workbench |
| **Backfill** | Post-deploy task | `cmd/backfill/` or task | Once, batch |
| **Config upgrade** | Command execution | CLI layer (`cli/`) | Per-machine, lazy |

**Config upgrades** are local file format changes (`.orc/config.json`). They:
- Run lazily on first command needing the config
- Live in CLI layer (may need DB access via wire)
- Must be idempotent and fail gracefully
- Cannot live in `config` package (no DB access)

## Git Hooks (Reminder-Based)

Git hooks provide **breadcrumbs**, not automation:

**post-checkout / post-merge:**
- Shows current branch
- Warns if schema may be out of sync
- Shows workbench DB status if present

**pre-commit:**
- Quality gate (enforced)
- Runs `make lint` and `make test`
