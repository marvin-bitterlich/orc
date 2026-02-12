# Config & Infrastructure

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
orc infra apply WORK-001    # Reconcile missing worktrees/configs (filesystem only)
orc infra cleanup           # Remove orphan directories (explicit action)
```

The infrastructure plan shows:
- **Workbenches**: Git worktrees for each workbench record
- **TMux**: Session and window state

**Important**: `infra apply` does NOT delete directories or manage tmux. It only reconciles missing worktrees and configs. Use `orc tmux apply` for tmux session management and `infra cleanup` to remove orphan directories.

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
| `orc infra apply` | No | Yes | No | Reconciles missing worktrees/configs (filesystem only) |
| `orc infra cleanup` | No | No | Yes | Removes orphan directories |
| `orc tmux connect` | No | No | No | Attaches to existing session |
