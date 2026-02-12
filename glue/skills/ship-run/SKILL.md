---
name: ship-run
description: |
  Launch shipment implementation by assigning tasks to IMP workers via Claude Teams.
  Use when user says "/ship-run", "run shipment", "start implementation", or "launch IMPs".
  Reads task dependency graph and proposes team shape before spawning workers.
---

# Ship Run

Launch shipment implementation by reading the task dependency graph, proposing a team shape, and spawning IMP workers via Claude Teams.

## Usage

```
/ship-run              (run focused shipment)
/ship-run SHIP-xxx     (run specific shipment)
```

## When to Use

- After /ship-plan has created tasks with dependencies
- When ready to begin parallel implementation
- Shipment must be in `ready` or `in-progress` status

## Flow

### Step 1: Get Shipment

If argument provided:
- Use specified SHIP-xxx

If no argument:
```bash
orc focus --show
```
- Use focused shipment
- If no focus: "No shipment focused. Run `orc focus SHIP-xxx` first."

### Step 2: Load Tasks and Dependencies

```bash
orc task list --shipment SHIP-xxx
```

For each task, check its dependencies:
```bash
orc task show TASK-xxx
```

Build the dependency graph:
- **Root tasks**: Tasks with no `depends_on` (can start immediately)
- **Dependent tasks**: Tasks blocked until their dependencies complete
- **Terminal tasks**: Tasks that nothing depends on

### Step 3: Preflight Checks

Run before spawning any workers:

```bash
# Working tree must be clean
git status --porcelain

# Project must compile/build
make dev 2>&1 || echo "BUILD FAILED"

# Tests must pass
make test 2>&1 || echo "TESTS FAILED"
```

If any check fails:
```
Preflight failed:
  [FAIL] <check description>

Fix issues before running. Failing state will propagate to all workers.
```

Do not proceed until all preflight checks pass.

### Step 4: Propose Team Shape

Analyze the dependency graph to determine optimal parallelism:

```
Task Dependency Graph:
  TASK-001: Move dev-only skills to repo-local
  TASK-002: Delete dead skills
  TASK-003: Add orphan cleanup (depends on: TASK-001, TASK-002)
  TASK-004: Update architecture.md (depends on: TASK-001, TASK-002)

Proposed team shape:

  Stream A (imp-alpha):  TASK-001 -> TASK-003
  Stream B (imp-bravo):  TASK-002 -> TASK-004

Workers needed: 2
```

**Rules for stream assignment:**
- Group tasks into streams by following dependency chains
- Independent root tasks go to separate streams (maximize parallelism)
- Dependent tasks follow their parent's stream when possible
- Minimize the number of workers (each has overhead)
- A single worker can handle 2-4 tasks sequentially

Present to Goblin:
```
Proposed team: 2 IMPs

  imp-alpha:  TASK-001 -> TASK-003
  imp-bravo:  TASK-002 -> TASK-004

Approve team shape? [y/n/edit]
```

Wait for explicit Goblin approval before proceeding.

### Step 5: Prepare Worker Context

For each IMP worker, build a context injection message containing:

1. **Mission**: Which tasks to work on, in order
2. **Spec context**: Summary note from the shipment (if exists)
3. **Dependency info**: Which tasks to complete before starting blocked tasks
4. **File paths**: Key files relevant to their tasks (from C2/C3 scope in task descriptions)

```
You are <imp-name>, an IMP worker on team <shipment-title>.

Your mission: <stream description> -- tasks <task-list>.

## How to work

1. Check TaskList to see available tasks
2. Claim a task with TaskUpdate (set owner, status to in_progress)
3. Do the work
4. Mark completed with TaskUpdate when done
5. Check TaskList for next available task
6. Message the team lead when blocked or done

## Important context

- Working directory: <workbench-path>
- Use `./orc` (local binary) for any ORC commands
- Use `make dev` to rebuild if needed, `make test` and `make lint` to verify
- Branch: <branch-name>

## Task details

<for each task in stream>
## Task #N: <title>
<description>
<depends-on info if any>
</for each>
```

### Step 6: Spawn Workers

Start each IMP as a Claude Teams worker, injecting the prepared context.

Report to Goblin:
```
Workers launched:
  imp-alpha: TASK-001, TASK-003 (2 tasks)
  imp-bravo: TASK-002, TASK-004 (2 tasks)

Monitor with: orc task list --shipment SHIP-xxx
```

### Step 7: Transition Shipment

If shipment is in `ready` status, transition to `in-progress`:
```bash
orc shipment status SHIP-xxx --set in-progress
```

## Guidelines

- **Always get Goblin approval** before spawning workers
- **Preflight is mandatory** -- never skip build/test checks
- **Prefer fewer workers** -- each worker has context overhead
- **Include spec note** in every worker's context injection
- **Tasks with no depends_on are parallelizable** -- assign to separate streams
- **Don't split fine-grained** -- let each worker handle a full stream

## Error Handling

| Error | Action |
|-------|--------|
| No tasks found | "No tasks in SHIP-xxx. Run /ship-plan first." |
| All tasks closed | "All tasks already closed. Nothing to run." |
| Preflight fails | Show failure, do not spawn workers |
| Circular dependency | "Circular dependency detected. Fix task graph." |
| Goblin rejects shape | Ask for edits, re-propose |

## Example Session

```
> /ship-run

[gets focused shipment SHIP-381]
[loads 6 tasks with dependencies]

Preflight checks:
  [PASS] Clean working tree
  [PASS] Build successful
  [PASS] Tests passing

Task Dependency Graph:
  TASK-998: Move dev-only skills to repo-local
  TASK-999: Delete dead skills
  TASK-1000: Add orphan cleanup (depends on: TASK-998, TASK-999)
  TASK-1001: Add --depends-on flag to task CLI
  TASK-1002: Update ship-plan (depends on: TASK-1001)
  TASK-1003: Create ship-run skill (depends on: TASK-1002)

Proposed team: 3 IMPs

  imp-alpha:   TASK-998  -> TASK-1000 -> TASK-1003
  imp-bravo:   TASK-1001 -> TASK-1002
  imp-charlie:  TASK-999

Approve team shape? [y/n/edit]
> y

[spawns 3 workers with context]

Workers launched:
  imp-alpha:   TASK-998, TASK-1000, TASK-1003 (3 tasks)
  imp-bravo:   TASK-1001, TASK-1002 (2 tasks)
  imp-charlie:  TASK-999 (1 task)

Monitor with: orc task list --shipment SHIP-381
```
