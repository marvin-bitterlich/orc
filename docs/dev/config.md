# Config, Actor Model & Infrastructure

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

### Place ID Prefixes

| Prefix | Actor Type | Role | Description |
|--------|------------|------|-------------|
| `BENCH-` | Workbench | Goblin or IMP | Workbench workspace (git worktree) |

### What NOT to store
- `commission_id` -- Trust the DB (workbench -> workshop -> factory -> commission)
- `current_focus` -- Stored in DB (`workbenches.focused_id`)

## Actor Model

ORC uses a two-actor model where agents operate from workbenches:

### Actors

| Actor | Role | Purpose |
|-------|------|---------|
| **Goblin** | Coordinator | Human's long-running workbench pane. Creates/manages ORC tasks with human. Memory and policy layer. |
| **IMP** | Worker | Disposable worker agent spawned by Claude Teams. Executes tasks using Teams primitives. |

The naming metaphors match the roles: Goblins are cunning coordinators, IMPs are industrious workers.

### Integration Model (Claude Teams)

ORC and Claude Teams have complementary roles:

| Layer | Owner | Responsibility |
|-------|-------|---------------|
| Memory & Policy | ORC (Goblin) | What to do and why |
| Execution | Teams (IMP) | How to do it and who does it |

- Goblin creates/manages ORC tasks with the human
- Teams workers (IMPs) execute using Teams primitives
- ORC provides context; Teams provides execution

### Key Relationships

- **Workshop -> Workbenches**: 1:many (a workshop contains multiple workbenches)
- **Goblin** lives in a workbench pane directly (no separate gatehouse)

### Summary as Shared Bus

The `orc summary` command serves as a shared information bus:
- Both IMPs and Goblins see the same data structure
- Filtering varies by role (IMP sees their shipments, Goblin sees all)
- Task children (plans) visible to both
- Status indicators show current state

## Workshop Commands

### set-commission
Sets the active commission for Goblin context. Must be run from a workshop workbench directory.

```bash
orc workshop set-commission COMM-001   # Set active commission
orc workshop set-commission --clear    # Clear active commission
```

The active commission is stored in `workshops.active_commission_id` and scopes Goblin's focus and operations.

## Infrastructure (Plan/Apply Pattern)

ORC separates **DB record creation** from **filesystem operations** using a plan/apply pattern.

### Key Principle

**DB commands create records only** -- they do NOT create directories, worktrees, or config files.

```bash
orc workbench create my-bench --workshop WORK-001   # Creates DB record only
```

### Infrastructure Workflow

Use `orc infra` to materialize physical infrastructure:

```bash
orc infra plan WORK-001     # Show what would be created/cleaned (dry run)
orc infra apply WORK-001    # Create infra + clean orphan tmux windows
orc infra cleanup           # Remove orphan directories (explicit action)
```

The infrastructure plan shows:
- **Workbenches**: Git worktrees for each workbench record
- **TMux**: Session and window state

**Important**: `infra apply` does NOT delete directories. It only creates infrastructure and removes orphan tmux windows (for archived workbenches). Use `infra cleanup` to explicitly remove orphan directories.

### TMux Connectivity

After infrastructure exists, connect to the session:

```bash
orc tmux connect WORK-001   # Attach to workshop's tmux session
```

### Command Reference

| Command | Creates DB Record | Creates Filesystem | Deletes Filesystem | Notes |
|---------|------------------|-------------------|-------------------|-------|
| `orc workbench create` | Yes | No | No | DB only |
| `orc commission start` | No | No | No | Starts tmux session only |
| `orc infra plan` | No | No | No | Shows what would change |
| `orc infra apply` | No | Yes | No | Creates infra, cleans tmux windows |
| `orc infra cleanup` | No | No | Yes | Removes orphan directories |
| `orc tmux connect` | No | No | No | Attaches to existing session |

### Why This Pattern?

1. **Predictability**: DB state is the source of truth; filesystem catches up via `apply`
2. **Idempotency**: `apply` is safe to run multiple times
3. **Visibility**: `plan` shows exactly what will change before committing
4. **Recovery**: If filesystem gets corrupted, `apply` reconstructs from DB
5. **Safety**: `apply` never deletes directories -- deletion requires explicit `cleanup` command
