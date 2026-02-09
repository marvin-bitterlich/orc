---
name: watchdog-monitor
description: Watchdog monitoring loop that watches over IMP. Checks should-continue, captures pane, detects state, and takes action.
---

# Watchdog Monitor

Monitoring loop for Watchdog agents. Periodically checks IMP state and takes corrective action when needed.

## Prerequisites

- Running as Watchdog (WATCHDOG_KENNEL_ID set)
- Patrol started (`orc patrol start BENCH-xxx`)
- Know the IMP's tmux target (session:window.2)

## Monitoring Loop

Run this loop until `orc shipment should-continue` returns exit code 1:

### Step 1: Check Should Continue

```bash
orc shipment should-continue
```

**Exit code 0**: Continue monitoring
**Exit code 1**: Stop monitoring (all tasks complete or not in auto mode)

If exit code 1, output:
```
Monitoring complete. IMP shipment finished or exited auto mode.
Ending patrol...
```
Then run `orc patrol end PATROL-xxx` and exit the skill.

### Step 2: Capture IMP Pane

Get the patrol target and capture content:

```bash
# Get patrol info
orc patrol status

# Capture last 50 lines of IMP pane
tmux capture-pane -t {target} -p -S -50
```

### Step 3: Detect State

Analyze the captured content to determine IMP state:

| State | Detection Pattern | Priority |
|-------|------------------|----------|
| **typed** | Input line has user text (not empty after prompt) | 1 (highest) |
| **menu** | Contains "?" or action selection prompt | 2 |
| **idle** | Contains "? for help" and no recent tool output | 3 |
| **error** | Contains "Error:", "error:", or stack traces | 4 |
| **working** | Recent tool output, file changes, or command execution | 5 (lowest) |

### Step 4: Take Action

Based on detected state:

| State | Action |
|-------|--------|
| **working** | No action - IMP is productive |
| **idle** | Send Enter key to nudge: `tmux send-keys -t {target} Enter` |
| **menu** | Send Escape to dismiss: `tmux send-keys -t {target} Escape` |
| **typed** | No action - user is typing, don't interrupt |
| **error** | Log the error, consider escalating if persistent |

### Step 5: Sleep and Repeat

```bash
sleep 30
```

Then return to Step 1.

## Example Session

```
> /watchdog-monitor

Starting monitoring loop for patrol PATROL-042...
Target: orc-dev:orc-044.2

[Loop 1]
Checking should-continue... continue=true (5 tasks remaining)
Capturing pane... 50 lines
Detecting state... working (recent file edits visible)
Action: none (IMP is working)
Sleeping 30s...

[Loop 2]
Checking should-continue... continue=true (5 tasks remaining)
Capturing pane... 50 lines
Detecting state... idle (? for help visible, no recent activity)
Action: nudging IMP (sending Enter)
Sleeping 30s...

[Loop 3]
Checking should-continue... continue=false (shipment complete)
Monitoring complete. IMP shipment finished.
Ending patrol PATROL-042...
Done.
```

## State Detection Details

### Detecting "typed"
Look for:
- Non-empty text after the last prompt character (> or $)
- Partial command that hasn't been executed

### Detecting "menu"
Look for:
- "Select an action" or similar menu prompts
- Numbered lists with action choices
- "Press Enter to" type prompts

### Detecting "idle"
Look for:
- "? for help" in the last few lines
- Long time since last tool output
- Repeated identical content across captures

### Detecting "error"
Look for:
- "Error:" or "error:" strings
- "Exception" or "failed"
- Stack traces with file:line patterns

## Error Handling

- If `orc patrol status` fails, the patrol may have ended externally. Exit gracefully.
- If tmux capture fails, the IMP pane may have closed. Report and end patrol.
- If nudge fails repeatedly (5+ times), escalate to Goblin.

## Notes

- 30-second interval balances responsiveness with overhead
- Never interrupt when user has typed text (respect human input)
- Watchdog should be silent when things are working
- Only log/output when taking action or when loop ends
