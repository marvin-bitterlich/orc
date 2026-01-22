# AGENTS.md - Development Rules for Claude Agents (ORC)

This file contains essential workflow rules for agents working on the ORC codebase.

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
   `internal/core/` contains domain logic (guards, FSM logic, planners). It must not import other ORC packages (stdlib only, plus other `core/` packages).

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

### FSM-First Development (for stateful entities)

Any entity with states/transitions (guards + lifecycle) requires an FSM spec in `specs/`.

**Workflow:**
1. Write/update the FSM spec (YAML) first
2. Derive/update the test matrix from the spec (manual until generators land)
3. Implement guards in `internal/core/<entity>/guards.go`
4. Implement service orchestration in `internal/app/<entity>_service.go`
5. Add repository tests for persistence correctness in `internal/adapters/sqlite/*_repo_test.go`

**Hard rule:**  
**Any new state-changing CLI/service method must map to an existing FSM transition or add one.**  
If you can’t point to a transition in `specs/<entity>-workflow.yaml`, the change is incomplete.

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

## Checklists

### Add Field to Entity

When adding a new field to an existing entity (e.g., adding `priority` to Task):

- [ ] Update model struct in `internal/models/<entity>.go`
- [ ] Update FSM spec if field affects state transitions (`specs/<entity>-workflow.yaml`)
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

### Add State/Transition to FSM

When adding a new state or transition to an entity’s state machine:

- [ ] Update FSM spec (`specs/<entity>-workflow.yaml`):
  - [ ] Add new state (if needed)
  - [ ] Add new event (if needed)
  - [ ] Add transition(s)
  - [ ] Define/adjust guards
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

- [ ] FSM spec in `specs/<entity>-workflow.yaml`
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

## Database Migrations (Atlas)

ORC uses [Atlas](https://atlasgo.io/) for declarative schema migrations. Atlas prevents FK reference corruption by validating the entire schema graph before applying changes.

### Why Atlas?

Hand-rolled SQLite migrations repeatedly caused FK reference corruption during table renames. Atlas catches these at validation time - you can't even *define* a schema with dangling FK references.

### Installation

```bash
brew install ariga/tap/atlas
```

### Core Workflow

**1. Inspect current schema:**
```bash
atlas schema inspect -u "sqlite:///$HOME/.orc/orc.db"
```

**2. Edit desired schema** in `schema.hcl` (declarative - say what you want, not how to get there)

**3. Preview migration:**
```bash
atlas schema diff \
  --from "sqlite:///$HOME/.orc/orc.db" \
  --to "file://schema.hcl" \
  --dev-url "sqlite://dev?mode=memory"
```

**4. Apply migration:**
```bash
atlas schema apply \
  --url "sqlite:///$HOME/.orc/orc.db" \
  --to "file://schema.hcl" \
  --dev-url "sqlite://dev?mode=memory"
```

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

## Common Mistakes to Avoid

❌ Writing business logic in adapters  
✅ Keep adapters as pure translation layers

❌ Importing adapters/infra from core/  
✅ Core has no non-core jmports

❌ Calling tmux directly from app  
✅ Use a port (adapter executes tmux)

❌ Skipping FSM spec for new stateful behavior  
✅ Spec → tests → implementation

❌ Claiming checks passed without running them  
✅ Run them and report explicitly

