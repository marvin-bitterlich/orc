# AGENTS.md - Development Rules for Claude Agents (ORC)

This file contains essential workflow rules for agents working on the ORC codebase.

## Development Setup

Run `make init` after cloning to install git hooks.

## Pre-Commit Checks (Enforced by Hook)

All commits must pass `make lint`, which runs:
- **check-test-presence**: Every service/repo/guard needs a test file
- **check-coverage**: Per-package coverage thresholds
- **schema-check**: Prevents hardcoded test schemas
- **golangci-lint**: Code quality
- **go-arch-lint**: Hex architecture boundaries

To bypass in emergencies: `git commit --no-verify` (will be audited)

---

## Config Files

ORC uses `.orc/config.json` for identity only (not state).

### Config Format (place_id model)

The config identifies "where you are" via a single `place_id`:

```json
{
  "version": "1.0",
  "place_id": "BENCH-014"
}
```

Or for Goblins:
```json
{
  "version": "1.0",
  "place_id": "GATE-003"
}
```

### Place ID Prefixes

| Prefix | Actor Type | Role | Description |
|--------|------------|------|-------------|
| `BENCH-` | Workbench | IMP | Implementation agent workspace (git worktree) |
| `GATE-` | Gatehouse | Goblin | Workshop gatekeeper (1:1 with workshop) |
| `WATCH-` | Watchdog | - | IMP monitor (1:1 with workbench) |

### What NOT to store
- ❌ `role` - Derived from place_id prefix (`BENCH-` → IMP, `GATE-` → Goblin)
- ❌ `commission_id` - Trust the DB (workbench → workshop → factory → commission)
- ❌ `current_focus` - Stored in DB (`workbenches.focused_id`)

---

## Actor Model

ORC uses a place-based actor model where identity is tied to "where you are":

### Actors

| Actor | Place | Role | Purpose |
|-------|-------|------|---------|
| IMP | Workbench (`BENCH-xxx`) | Implementation | Executes tasks, writes code |
| Goblin | Gatehouse (`GATE-xxx`) | Review | Reviews plans, handles escalations |
| Watchdog | Watchdog (`WATCH-xxx`) | Monitor | Monitors IMP progress |

### Key Relationships

- **Workshop → Gatehouse**: 1:1 (every workshop has exactly one gatehouse)
- **Workbench → Watchdog**: 1:1 (every workbench has exactly one watchdog)
- **Workshop → Workbenches**: 1:many (a workshop contains multiple workbenches)

### Connect Command

`orc connect` establishes agent identity. It validates role against place:

```bash
# From a workbench directory (has .orc/config.json with place_id=BENCH-xxx)
orc connect                # Auto-detects IMP role
orc connect --role imp     # Explicit IMP role (validates against place)
orc connect --role goblin  # ERROR: Goblin role not allowed from workbench

# From a gatehouse directory (has .orc/config.json with place_id=GATE-xxx)
orc connect                # Auto-detects Goblin role
orc connect --role goblin  # Explicit Goblin role
orc connect --role imp     # ERROR: IMP role not allowed from gatehouse
```

### Summary as Shared Bus

The `orc summary` command serves as a shared information bus:
- Both IMPs and Goblins see the same data structure
- Filtering varies by role (IMP sees their shipments, Goblin sees all)
- Task children (plans, approvals, escalations, receipts) visible to both
- Status indicators: ✓ (approved/complete), ⚠ (escalated/pending)

---

## Workshop Commands

### set-commission
Sets the active commission for Goblin context. Must be run from a workshop gatehouse directory.

```bash
orc workshop set-commission COMM-001   # Set active commission
orc workshop set-commission --clear    # Clear active commission
```

The active commission is stored in `workshops.active_commission_id` and scopes Goblin's focus and operations.

---

## Infrastructure (Plan/Apply Pattern)

ORC separates **DB record creation** from **filesystem operations** using a plan/apply pattern.

### Key Principle

**DB commands create records only** — they do NOT create directories, worktrees, or config files.

```bash
orc workbench create my-bench --workshop WORK-001   # Creates DB record only
orc workbench like feature-v2                        # Creates DB record only
```

### Infrastructure Workflow

Use `orc infra` to materialize physical infrastructure:

```bash
orc infra plan WORK-001     # Show what would be created (dry run)
orc infra apply WORK-001    # Create directories, worktrees, configs
```

The infrastructure plan shows:
- **Gatehouse**: Workshop coordination directory (`~/.orc/ws/WORK-xxx-slug/`)
- **Workbenches**: Git worktrees for each workbench record

### TMux Connectivity

After infrastructure exists, connect to the session:

```bash
orc tmux connect WORK-001   # Attach to workshop's tmux session
```

### Command Reference

| Command | Creates DB Record | Creates Filesystem | Notes |
|---------|------------------|-------------------|-------|
| `orc workbench create` | ✓ | ✗ | DB only |
| `orc workbench like` | ✓ | ✗ | DB only |
| `orc commission start` | ✗ | ✗ | Starts tmux session only |
| `orc infra plan` | ✗ | ✗ | Shows what would change |
| `orc infra apply` | ✗ | ✓ | Creates gatehouse, worktrees, configs |
| `orc tmux connect` | ✗ | ✗ | Attaches to existing session |

### Why This Pattern?

1. **Predictability**: DB state is the source of truth; filesystem catches up via `apply`
2. **Idempotency**: `apply` is safe to run multiple times
3. **Visibility**: `plan` shows exactly what will change before committing
4. **Recovery**: If filesystem gets corrupted, `apply` reconstructs from DB

---

## Build & Development

**ALWAYS use the Makefile for building and installing ORC:**

```bash
make dev        # Build local ./orc for development (preferred)
make install    # Build and install globally with local-first shim
make test       # Run all tests
make lint       # Run golangci-lint + architecture linting (go-arch-lint)
make clean      # Clean build artifacts
```

### Binary Management Convention

When developing ORC itself, **always use `./orc`** (the local binary):

```bash
make dev
./orc status
./orc help
```

**Why this matters:**
- The local-first shim prefers `./orc` when present
- Ensures you're testing your actual changes
- Prevents confusion between global and development binaries
- `make dev && ./orc <cmd>` is the canonical development workflow

---

## Architecture Rules

ORC follows a hexagonal (ports & adapters) architecture with strict layer boundaries.

**The architecture linter config (`.go-arch-lint.yml`) is the source of truth.**  
If this document and the linter disagree, **the linter wins**.

### Layer Hierarchy (intent)

```
┌─────────────────────────────────────────────────────────┐
│                        cmd/                             │
│                   (entry points)                        │
├─────────────────────────────────────────────────────────┤
│                      cli/                               │
│              (Cobra commands, thin)                     │
├─────────────────────────────────────────────────────────┤
│                      wire/                              │
│           (dependency injection only)                   │
├─────────────────────────────────────────────────────────┤
│                      app/                               │
│     (orchestration: uses ports, no direct I/O)          │
├─────────────────────────────────────────────────────────┤
│                      ports/                             │
│               (interfaces only)                         │
├─────────────────────────────────────────────────────────┤
│                      core/                              │
│        (pure domain logic, no dependencies)             │
└─────────────────────────────────────────────────────────┘

adapters/ implements ports/ and performs I/O (SQLite, tmux, filesystem, etc.)
```

### Architecture Principles

1. **Core is pure**  
   `internal/core/` contains domain logic (guards, planners). It must not import other ORC packages (stdlib only, plus other `core/` packages).

2. **Ports are contracts**  
   `internal/ports/` contains interfaces only (stdlib only).

3. **App orchestrates**  
   `internal/app/` coordinates workflow using `core` + `ports` + `models`. It must not reach into infrastructure packages directly.

4. **Adapters are boring**  
   `internal/adapters/` contains translation and I/O only. No business logic (no ID generation, no default statuses, no transition semantics).

5. **CLI is thin**  
   `internal/cli/` commands parse args, call app services via ports/wire, and render output. They must not orchestrate workflows.

6. **tmux is an adapter concern**  
   `internal/app` must not import `internal/tmux`. Access tmux via a port (or via effect execution through a port).

Run `make lint` to verify architecture compliance.

---

## Testing Rules

### Table-Driven Tests (Default Pattern)

Default to table-driven tests for guards, planners, validation, and service decision logic.

```go
func TestCanPauseTask(t *testing.T) {
    tests := []struct {
        name        string
        ctx         StatusTransitionContext
        wantAllowed bool
        wantReason  string
    }{
        {
            name: "can pause in_progress task",
            ctx:  StatusTransitionContext{TaskID: "TASK-001", Status: "in_progress"},
            wantAllowed: true,
        },
        {
            name: "cannot pause ready task",
            ctx:  StatusTransitionContext{TaskID: "TASK-001", Status: "ready"},
            wantAllowed: false,
            wantReason:  "can only pause in_progress tasks (current status: ready)",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CanPauseTask(tt.ctx)
            if result.Allowed != tt.wantAllowed {
                t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
            }
            if !tt.wantAllowed && result.Reason != tt.wantReason {
                t.Errorf("Reason = %q, want %q", result.Reason, tt.wantReason)
            }
        })
    }
}
```

### Test Pyramid

```
┌─────────────────────────────┐
│     Integration Tests       │  ← Sparse: end-to-end CLI flows / wiring
├─────────────────────────────┤
│     Repository Tests        │  ← Medium: SQL correctness + persistence invariants
├─────────────────────────────┤
│      Service Tests          │  ← Most: orchestration logic (mock ports)
├─────────────────────────────┤
│       Guard Tests           │  ← Foundation: pure functions
└─────────────────────────────┘
```

### Test Helpers

Use `testutil_test.go` helpers in `internal/adapters/sqlite/` where available to avoid repeating DB setup + seeding.

---

## Verification Discipline

LLMs are prone to skipping checks. ORC’s workflow requires explicit verification.

### Plans Must Include Checks

Every implementation plan must explicitly list:
- [ ] Tests to run
- [ ] Lint checks to pass
- [ ] Manual verification steps (if applicable)

### Completion Must Report What Ran (and what didn’t)

When completing work, report verification explicitly:

```
✅ Ran: make test (all passing)
✅ Ran: make lint (no issues)
⏭️ Skipped: <check> (reason)
```

**Rule:** If a check was not run, it must be explicitly marked as skipped with a reason. Never imply success.

---

## Test Commission for CLI Validation

When developing changes that affect CLI display (summary, containers, leafs, etc.), use the test commission to validate output:

```bash
./orc summary --commission COMM-003
```

### What's in COMM-003

The test commission contains representative examples of all container types:
- **Shipment** (SHIP-205) with cycles, work orders, plans, and tasks (including paused status)
- **Conclave** (CON-007) with nested tome and notes (including pinned note)
- **Tome** (TOME-008, standalone) with notes

Also includes items with various statuses (ready, paused, draft, complete) and pinned items.

### When to Use

Run `orc summary --commission COMM-003` after changes to:
- Summary display logic
- Container creation/update commands
- Leaf item (task, note, plan) display
- Status filtering or colorization
- Pinned item display
- Hierarchical nesting

### Maintenance

If you add new container types or display features, add corresponding test data to COMM-003.

---

## Checklists

### Add Field to Entity

When adding a new field to an existing entity (e.g., adding `priority` to Task):

- [ ] Update model struct in `internal/models/<entity>.go`
- [ ] Update SQL schema in `internal/db/schema.sql`
- [ ] Create migration in `internal/db/migrations/`
- [ ] Update repository:
  - [ ] `internal/adapters/sqlite/<entity>_repo.go`
  - [ ] `internal/adapters/sqlite/<entity>_repo_test.go`
- [ ] Update service if field has business logic:
  - [ ] `internal/app/<entity>_service.go`
  - [ ] `internal/app/<entity>_service_test.go`
- [ ] Update CLI if field is user-facing
- [ ] Run: `make test && make lint`

### Add State/Transition

When adding a new state or transition to an entity's state machine:

- [ ] Update core guards + tests:
  - [ ] `internal/core/<entity>/guards.go`
  - [ ] `internal/core/<entity>/guards_test.go`
- [ ] Update service + tests:
  - [ ] `internal/app/<entity>_service.go`
  - [ ] `internal/app/<entity>_service_test.go`
- [ ] Update CLI if user-triggerable
- [ ] Run: `make test && make lint`

### Add CLI Command

- [ ] Create: `internal/cli/<command>.go`
- [ ] Keep it thin: parse args/flags, call services, render output
- [ ] Inject dependencies via wire (no globals)
- [ ] Manual smoke: `make dev && ./orc <command> --help`
- [ ] Run: `make test && make lint`

### Add New Entity (with Repository)

When adding a new entity that requires persistence (e.g., CycleWorkOrder, Receipt):

- [ ] Guards in `internal/core/<entity>/guards.go`
- [ ] Guard tests in `internal/core/<entity>/guards_test.go`
- [ ] Schema in `internal/db/schema.go`
- [ ] Migration in `internal/db/migrations/`
- [ ] Secondary port interface in `internal/ports/secondary/persistence.go`
- [ ] Primary port interface in `internal/ports/primary/<entity>.go`
- [ ] **Repository implementation + tests** (REQUIRED):
  - [ ] `internal/adapters/sqlite/<entity>_repo.go`
  - [ ] `internal/adapters/sqlite/<entity>_repo_test.go`
- [ ] Service implementation in `internal/app/<entity>_service.go`
- [ ] CLI commands in `internal/cli/<entity>.go`
- [ ] Wire registration in `internal/wire/`
- [ ] Run: `make test && make lint`

**Hard rule:** Repository tests are NOT optional. Every `*_repo.go` MUST have a corresponding `*_repo_test.go`.

---

## Data & Config Changes

**Don't overload "migration"** — these are distinct operations:

| Term | When | Where | Runs |
|------|------|-------|------|
| **Schema migration** | Deploy | `internal/db/schema.go` | Once, centralized |
| **Backfill** | Post-deploy task | `cmd/backfill/` or task | Once, batch |
| **Config upgrade** | Command execution | CLI layer (`cli/`) | Per-machine, lazy |

**Config upgrades** are local file format changes (`.orc/config.json`). They:
- Run lazily on first command needing the config
- Live in CLI layer (may need DB access via wire)
- Must be idempotent and fail gracefully
- Cannot live in `config` package (no DB access)

---

## Database Migrations (Atlas)

ORC uses [Atlas](https://atlasgo.io/) for declarative schema migrations. Atlas prevents FK reference corruption by validating the entire schema graph before applying changes.

### Why Atlas?

Hand-rolled SQLite migrations repeatedly caused FK reference corruption during table renames. Atlas catches these at validation time - you can't even *define* a schema with dangling FK references.

### Installation

```bash
brew install ariga/tap/atlas
```

### Core Workflow

**CRITICAL: Always use `--exclude` for SQLite autoindexes.**

SQLite auto-generates indexes named `sqlite_autoindex_<table>_N` for UNIQUE and PRIMARY KEY constraints. Atlas has a bug where it generates incorrect names when trying to drop these, causing migrations to fail with "no such index" errors.

**1. Inspect current schema:**
```bash
atlas schema inspect \
  -u "sqlite:///$HOME/.orc/orc.db" \
  --exclude "*.sqlite_autoindex*[type=index]"
```

**2. Edit desired schema** in `schema_inspected.hcl` (declarative - say what you want, not how to get there)

**3. Preview migration:**
```bash
atlas schema diff \
  --from "sqlite:///$HOME/.orc/orc.db" \
  --to "file://schema_inspected.hcl" \
  --dev-url "sqlite://dev?mode=memory" \
  --exclude "*.sqlite_autoindex*[type=index]"
```

**4. Apply migration:**
```bash
atlas schema apply \
  --url "sqlite:///$HOME/.orc/orc.db" \
  --to "file://schema_inspected.hcl" \
  --dev-url "sqlite://dev?mode=memory" \
  --exclude "*.sqlite_autoindex*[type=index]"
```

**Why the `--exclude` flag?**
- SQLite manages `sqlite_autoindex_*` indexes internally for unique constraints
- Atlas inspect captures them, but diff generates wrong drop statements (e.g., `DROP INDEX tags_name` instead of `DROP INDEX sqlite_autoindex_tags_2`)
- The exclude pattern tells Atlas to ignore these on both sides of the diff
- This is an Atlas bug (confirmed with canary v1.0.1), not a config issue

### Key Behaviors

- **FK validation**: Atlas refuses to process schemas with dangling FK references
- **Auto-generates SQL**: Handles SQLite's rename-recreate dance automatically
- **Dependency ordering**: Knows which tables reference which, applies changes in safe order
- **Data preservation**: Copies data during table recreates

### Example: Renaming a Table

If you rename `users` → `accounts`, Atlas will:
1. `PRAGMA foreign_keys = off`
2. Recreate any tables with FKs pointing to `users` (updating refs to `accounts`)
3. Copy data
4. Create `accounts`
5. `PRAGMA foreign_keys = on`

You just declare the end state. Atlas figures out the migration path.

### Golden Rule

**Never write migration SQL by hand.** Declare the desired schema, let Atlas diff and apply.

---

## Creating Containers

### Tomes
- Use `--conclave CON-xxx` to create in a specific conclave
- Tomes without a conclave exist at commission root (orphan tomes)
- `orc tome unpark TOME-xxx --conclave CON-xxx` moves tome to a Conclave

### Shipments
- Use `--conclave CON-xxx` to create in a specific conclave
- Use `--shipyard` to create in the commission's Shipyard
- `orc shipment park SHIP-xxx` moves to Shipyard
- `orc shipment unpark SHIP-xxx --conclave CON-xxx` moves to Conclave

---

## Common Mistakes to Avoid

❌ Writing business logic in adapters
✅ Keep adapters as pure translation layers

❌ Importing adapters/infra from core/
✅ Core has no non-core imports

❌ Calling tmux directly from app
✅ Use a port (adapter executes tmux)

❌ Claiming checks passed without running them
✅ Run them and report explicitly

❌ Running Atlas without `--exclude "*.sqlite_autoindex*[type=index]"`
✅ Always include the exclude flag for SQLite migrations (see Atlas section above)

