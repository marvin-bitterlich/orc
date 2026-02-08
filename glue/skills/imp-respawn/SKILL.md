---
name: imp-respawn
description: Respawn the current tmux pane for a clean IMP context. Shows focus and pane command, asks for confirmation, then restarts the pane.
---

# IMP Respawn

Respawn the current tmux pane to get a fresh IMP session with clean context.

## Usage

```
/imp-respawn
```

## Flow

### Step 1: Check Pane Root Command

```bash
tmux display-message -p '#{pane_start_command}'
```

**Guard**: If empty or missing, output error and STOP:
```
Error: No pane start command found.

This pane wasn't set up by ORC infrastructure. Respawning would leave
you with an empty shell instead of a fresh IMP session.

Set up the workbench properly with: orc infra apply WORK-xxx
```

### Step 2: Get Current Focus

```bash
orc status
```

Extract the focused entity (SHIP-xxx, COMM-xxx, etc.) and title.

If no focus set, note this but continue (valid to respawn between shipments).

### Step 3: Display Summary and Confirm

Output this format exactly:

```
## imp-respawn

Current focus: SHIP-xxx (title)
Pane root command: orc connect

On respawn:
  1. This session terminates immediately
  2. Pane restarts with: orc connect
  3. Fresh Claude session begins
  4. New IMP runs: orc prime

Respawn now? [y/n]
```

If no focus:
```
## imp-respawn

Current focus: No focus set
Pane root command: orc connect

On respawn:
  1. This session terminates immediately
  2. Pane restarts with: orc connect
  3. Fresh Claude session begins
  4. New IMP runs: orc prime

Respawn now? [y/n]
```

### Step 4: Handle Response

If user responds **y** or **yes**:
```bash
tmux respawn-pane -k -t "$TMUX_PANE"
```

This immediately terminates the current session. No success message is possible.

If user responds **n** or **no**:
```
Respawn cancelled.
```

## Notes

- The skill cannot report success because respawn kills the process immediately
- This is expected behavior - the user will see their terminal restart
- Use this after completing or planning a shipment to get clean context
