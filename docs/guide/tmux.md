# TMux Integration

**Status**: Living document
**Last Updated**: 2026-02-08

ORC uses TMux for multi-agent session management. Each workshop has a dedicated TMux session with windows for workbenches.

## Window/Pane Layouts

### IMP Layout (Workbench Window)

Standard 3-pane layout for implementation work:

```
┌─────────────────┬─────────────────┐
│                 │ claude (IMP)    │
│ vim editor      ├─────────────────┤
│ (left pane)     │ shell           │
│                 │                 │
└─────────────────┴─────────────────┘
```

| Pane | Purpose | Respawn Command |
|------|---------|-----------------|
| 1 (left) | Editor | `vim` |
| 2 (top-right) | Claude IMP | `orc connect` |
| 3 (bottom-right) | Shell | (none) |

### Goblin Layout (Coordination Window)

The Goblin (coordinator) lives in a workbench pane directly:

```
┌─────────────────────┬──────────────┐
│                     │   vim        │
│      claude         │──────────────│
│    (Goblin)         │   shell      │
│                     │              │
└─────────────────────┴──────────────┘
```

| Pane | Purpose | Respawn Command |
|------|---------|-----------------|
| 1 (left) | Claude Goblin | `orc connect` |
| 2 (top-right) | Editor | `vim` |
| 3 (bottom-right) | Shell | (none) |

### How Layouts Are Created

Layouts are created by `orc infra apply WORK-xxx`:

1. Creates TMux session for workshop
2. Creates windows for each workbench
3. Sets respawn commands so panes can be refreshed

## Pane Respawn

Panes store their "start command" for respawning. This enables:
- Fresh context window for Claude (clears conversation)
- Recovery after pane crash
- Consistent state across sessions

**To respawn a pane:**
1. Right-click on pane in TMux
2. Select "Respawn Pane"
3. Pane restarts with its original command

The start command is stored via `respawn-pane -k` during window creation.

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
- Shows current focus (→ SHIP-xxx)
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

### Create Fresh Claude Context

1. Go to the Claude pane (pane 2)
2. Right-click → Respawn Pane
3. Claude restarts with fresh conversation

Alternatively, kill the pane and let it respawn:
```
prefix+x (kill pane, then confirm)
```
Then respawn with start command.

### Switch Between Workshops

1. `prefix+S` - Open ORC session picker
2. Navigate to desired workshop/window
3. Press Enter to switch

### View Workshop State

1. Double-click window in statusline
2. Or: `prefix+:` then `display-popup -w 100 -h 30 'orc summary | less'`

## Next Steps

- [docs/guide/common-workflows.md](common-workflows.md) - IMP/Goblin workflows
- [docs/reference/glue.md](../reference/glue.md) - Skills that interact with TMux
- [docs/reference/architecture.md](../reference/architecture.md) - System design
