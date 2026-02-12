# TMux Integration

**Status**: Living document
**Last Updated**: 2026-02-12

ORC uses TMux with [gotmux](https://github.com/GianlucaP106/gotmux) for programmatic multi-agent session management. Each workshop has a dedicated TMux session with windows for workbenches.

## Workflow Overview

The ORC tmux workflow follows this sequence:

1. **Create workbench** -- `orc workbench create` (creates DB record + worktree + config)
2. **Apply tmux session** -- `orc tmux apply WORK-xxx [--yes]` (creates/reconciles session)
3. **Connect** -- `orc tmux connect WORK-xxx` (attaches to session)

The `apply` command is the single entry point for all tmux session management. It compares
desired state (from the database) with actual tmux state and reconciles the difference.

## Gotmux Integration

ORC uses gotmux's Go API for programmatic tmux session management. Sessions are created directly via the library, not via YAML config files.

**Key operations:**
- `orc tmux apply WORK-xxx` - creates/reconciles session (plan + confirm)
- `orc tmux apply WORK-xxx --yes` - creates/reconciles session (immediate)
- `orc tmux connect WORK-xxx` - attaches to existing session
- `orc tmux enrich WORK-xxx` - re-applies enrichment (idempotent)
- `tmux kill-session -t WORK-xxx` - stops session (standard tmux)

**Benefits:**
- Direct Go API control (no YAML generation)
- Type-safe session manipulation
- Programmatic pane creation with pane options
- Plan/apply pattern for safe reconciliation
- Maintained library with stable API

## Window/Pane Layouts

### Standard Workbench Window

Default 3-pane layout created by gotmux:

```
+-----------------------+----------------+
|                       |  goblin        |
|      vim              +----------------+
|                       |  shell         |
+-----------------------+----------------+
```

| Pane | Role (@pane_role) | Purpose |
|------|-------------------|---------|
| 0 (left) | `vim` | Editor (main-vertical layout, 50% width) |
| 1 (top-right) | `goblin` | Coordinator pane |
| 2 (bottom-right) | `shell` | Shell |

### Guest Panes

**Guest panes** are any panes without a `@pane_role` tmux pane option. They are typically created by:
- Claude Teams spawning IMP workers (`break-pane` from existing pane)
- Manual pane splits by the user

Guest panes are first-class citizens - ORC accommodates them rather than forcing them into a predetermined layout.

Run `orc tmux apply WORK-xxx --yes` to relocate guest panes to a dedicated `-imps` window.

## Pane Identity Model

ORC uses tmux pane options for pane identity:

- **`@pane_role`** (authoritative) -- set via `pane.SetOption("@pane_role", role)` at creation. Readable by tmux format strings (`#{@pane_role}`). Used by ORC to identify panes.
- **`@bench_id`** -- workbench ID (e.g., `BENCH-001`). Provides workbench context.
- **`@workshop_id`** -- workshop ID (e.g., `WORK-001`). Provides workshop context.

| Role | Description |
|------|-------------|
| `vim` | Editor pane (always exists) |
| `goblin` | Coordinator pane (primary workbench only) |
| `shell` | Shell pane (always exists) |
| (unset) | Guest pane (Claude Teams workers or manual splits) |

**Note**: Shell environment variables (`PANE_ROLE`, etc.) are NOT used. All identity is via tmux pane options which are readable by tmux format strings.

## TMux Commands

### orc tmux apply

Reconciles tmux session state for a workshop. This is the single command for creating,
updating, and maintaining sessions.

```bash
orc tmux apply WORK-xxx          # Show plan, prompt for confirmation
orc tmux apply WORK-xxx --yes    # Apply immediately
```

**What it does:**
1. Compares desired state (workbenches from DB) with actual tmux state
2. Creates session if it doesn't exist
3. Adds windows for missing workbenches
4. Relocates guest panes to -imps windows
5. Prunes dead panes in -imps windows
6. Kills empty -imps windows (all panes dead)
7. Reconciles layout (main-vertical, 50% main-pane-width)
8. Applies ORC enrichment (bindings, pane titles)

### orc tmux connect

Attaches to an existing workshop session.

```bash
orc tmux connect WORK-xxx
```

### orc tmux enrich

Applies ORC-specific tmux bindings and enrichment to the current or specified session.

```bash
orc tmux enrich              # Enrich detected workshop session
orc tmux enrich WORK-xxx     # Enrich specific workshop
```

**What it does:**
- Configures mouse support
- Sets up ORC key bindings (session picker, summary popup, etc.)
- Applies statusline formatting
- Enables context display

Note: `orc tmux apply` runs enrichment automatically. Use `enrich` only when you need to
re-apply enrichment without full reconciliation.

## Session Management

ORC creates tmux sessions programmatically via the gotmux Go library. There are no configuration files - sessions are created directly from DB state.

To inspect or modify sessions, use standard tmux commands:
```bash
tmux list-sessions                    # Show all sessions
tmux list-windows -t WORK-xxx         # Show windows in session
tmux kill-session -t WORK-xxx         # Stop session
```

## Session Browser

### Standard Session Browser (prefix+s)

Press `prefix+s` (default: Ctrl-b, then s) to open TMux's session browser.

Format shows ORC context:
```
session1 [WORK-001] - Project Name [COMM-001]
session2 [WORK-002] - Other Project [COMM-002]
```

### ORC Session Picker (prefix+S)

Press `prefix+S` (capital S) for enhanced ORC picker:

- Vertical split with list and preview
- Shows agent type (IMP/GOBLIN) per window
- Shows current focus (-> SHIP-xxx)
- Live preview of selected pane content

Navigation:
- `j`/`k` or arrows to move
- `Enter` or `l` to select
- `q` to cancel

## Summary Menu

### Double-Click on Window

Double-click a window name in the statusline to open `orc summary` in a popup:

- Shows commission tree
- 100 columns x 30 rows
- Press `q` to close

### Right-Click Context Menu

Right-click the statusline for a context menu:

| Option | Action |
|--------|--------|
| Show Summary | `orc summary \| less` |
| Archive Workbench | `orc infra archive-workbench` |
| Swap Window | Standard TMux swap |
| Mark Pane | Standard TMux mark |
| Kill Window | Standard TMux kill |
| Respawn Pane | Restart pane with start command |
| Rename Window | Standard TMux rename |
| New Window | Standard TMux new window |

## Session Click

Click on the session name in the statusline (left side) to open the session browser. This is equivalent to `prefix+s`.

## Environment Variables

ORC sets these environment variables on TMux sessions:

| Variable | Scope | Purpose |
|----------|-------|---------|
| `ORC_WORKSHOP_ID` | Session | Workshop ID (WORK-xxx) |
| `ORC_CONTEXT` | Session | Active commissions summary |

These enable the session browser to show ORC context.

## Window Options

ORC tracks agent state via TMux window options:

| Option | Example | Purpose |
|--------|---------|---------|
| `@orc_agent` | `IMP-main@BENCH-001` | Agent identity |
| `@orc_focus` | `SHIP-334: Docs overhaul` | Current focus |

These enable the ORC session picker to show agent details.

## Common Operations

### Connect to Workshop

```bash
orc tmux connect WORK-xxx
```

Finds and attaches to the workshop's TMux session.

### Start a New Workshop Session

```bash
# 1. Create workshop/workbench (if needed)
orc workbench create --workshop WORK-001 --repo-id REPO-001

# 2. Apply tmux session (creates + enriches in one step)
orc tmux apply WORK-001 --yes

# 3. Attach
orc tmux connect WORK-001
```

### Manage Guest Panes

When Claude Teams spawns IMP workers or you manually split panes:

```bash
# Apply reconciles everything: relocates guest panes, prunes dead panes, etc.
orc tmux apply WORK-001 --yes
```

### Stop a Window

```bash
tmux kill-window -t "Workshop Name:bench-1"
```

Stops a specific workbench window while leaving the rest of the session running.

### Switch Between Workshops

1. `prefix+S` - Open ORC session picker
2. Navigate to desired workshop/window
3. Press Enter to switch

### View Workshop State

1. Double-click window in statusline
2. Or: `prefix+:` then `display-popup -w 100 -h 30 'orc summary | less'`

## Next Steps

- [docs/common-workflows.md](common-workflows.md) - IMP/Goblin workflows
- [docs/dev/glue.md](dev/glue.md) - Skills that interact with TMux
- [docs/architecture.md](architecture.md) - System design
