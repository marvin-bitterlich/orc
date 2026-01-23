# Plan Task Command

Lightweight task planning - finds the next ready task and plans it.

## Process

### Step 1: Get Current Focus

```bash
./orc status
```

Note the **focused Shipment ID**. If no shipment is focused, stop and ask El Presidente to focus one first with `orc focus SHIP-XXX`.

### Step 2: Find Next Task

```bash
./orc task list --shipment SHIP-XXX --status ready
```

Pick the **first task** (lowest ID = next in queue).

If no ready tasks, report "All tasks complete for this shipment."

### Step 3: Claim the Task

```bash
./orc task claim TASK-XXX
```

This marks the task as `implement` (in progress).

### Step 4: Get Task Details

```bash
./orc task show TASK-XXX
```

Read the full description to understand the scope.

### Step 5: Read AGENTS.md

```bash
cat AGENTS.md
```

**You MUST read AGENTS.md before planning.** Your plan must follow its architecture rules.

### Step 6: Enter Plan Mode

Use the `EnterPlanMode` tool now.

Design your plan to complete the task. Include:
- Files to modify
- Specific changes
- Verification steps

### Step 7: Exit Plan Mode

When ready, use `ExitPlanMode` to present for approval.

---

## After Approval

Implement the plan, then mark the task complete:

```bash
./orc task complete TASK-XXX
```

Run `/plan-task` again for the next one.

---

## Task Status Flow

```
ready → claim → implement → complete → complete
```

- `ready`: Available to work on
- `implement`: Claimed, work in progress
- `complete`: Done
