# Integration Tests

ORC has four integration-level validation skills that test the system beyond unit tests. This document maps what each covers, how they relate, and where the gaps are.

## Coverage Map

### orc-self-test

**Skill location:** `.claude/skills/orc-self-test/`
**Invocation:** `/orc-self-test`
**Automated:** Yes (agent-driven, no human interaction needed)

Tests the plan/apply infrastructure pattern end-to-end:

| What | Verified |
|------|----------|
| Factory CRUD | create, delete |
| Workshop CRUD | create, archive, delete |
| Workbench CRUD | create, archive, delete |
| Infra plan | Shows correct CREATE/MISSING states |
| Infra apply | Creates filesystem dirs, config files |
| TMux sessions | Session created, windows present |
| Archive lifecycle | Archive triggers infra cleanup on next apply |
| Cleanup | No orphan dirs or tmux sessions after delete |

**Not covered:** Commissions, shipments, tasks, notes, tomes, plans, focus, status transitions.

### bootstrap-test

**Skill location:** `.claude/skills/bootstrap-test/`
**Invocation:** `/bootstrap-test`
**Automated:** Yes (script runs in macOS VM via Tart)

Tests `make bootstrap` on a completely fresh system:

| What | Verified |
|------|----------|
| Makefile bootstrap target | Completes without error |
| Go installation | Installs via Homebrew |
| orc binary | In PATH and responds to commands |
| FACT-001 | Default factory created |
| REPO-001 | ORC repo registered with correct path |
| Basic CLI smoke | Commission create, workshop create, summary, doctor |

**Not covered:** `orc bootstrap` command (Phase 3), first-run skill, shipment/task workflows.

#### Requirements

- **Apple Silicon Mac** — Uses Virtualization.framework (no Intel support)
- **tart** — macOS VM manager: `brew install cirruslabs/cli/tart`
- **sshpass** — Non-interactive SSH: `brew install sshpass`

#### Running

```bash
make bootstrap-test
```

Or directly with options:

```bash
./scripts/bootstrap-test.sh --verbose           # Show detailed progress
./scripts/bootstrap-test.sh --keep-on-failure   # Keep VM for debugging if test fails
./scripts/bootstrap-test.sh --shell             # Bootstrap then drop into VM shell
```

#### What It Does

1. Creates a fresh macOS Tahoe VM
2. Installs Go via Homebrew
3. Copies ORC repo into VM
4. Runs `make bootstrap`
5. Verifies `orc` is in PATH and works
6. Cleans up VM on success

**First run note:** The first run will auto-pull the macOS base image (~25GB). Subsequent runs reuse the cached image. Typical run: ~70-80 seconds.

#### Debugging Failures

```bash
./scripts/bootstrap-test.sh --keep-on-failure --verbose
```

Then SSH into the VM:

```bash
ssh admin@$(tart ip orc-bootstrap-test-XXXX)
# Password: admin
```

To clean up afterward:

```bash
tart stop orc-bootstrap-test-XXXX
tart delete orc-bootstrap-test-XXXX
```

### bootstrap-exercise

**Skill location:** `.claude/skills/bootstrap-exercise/`
**Invocation:** `/bootstrap-exercise`
**Automated:** Semi-manual (agent creates test factory, human walks through first-run)

Tests the `orc bootstrap` → `/orc-first-run` skill chain on the dev machine:

| What | Verified |
|------|----------|
| orc bootstrap command | Launches Claude with correct directive |
| orc-first-run skill | Creates commission, workshop, workbench |
| --factory flag | Isolates test to dedicated factory |
| Entity creation via skill | Commission, workshop, workbench exist after flow |
| REPO-001 stability | Still exists after exercise |
| Cleanup | Archive pattern works for test entities |

**Not covered:** Shipment creation, task workflows, note-taking, the ship-* skill chain.

### docs-doctor

**Skill location:** `.claude/skills/docs-doctor/`
**Invocation:** `/docs-doctor`
**Automated:** Yes (agent-driven with parallel haiku subagents)

Validates documentation accuracy against code reality:

| What | Verified |
|------|----------|
| Internal links | All markdown cross-references resolve |
| Lane rules | README has no agent instructions, CLAUDE.md has no human onboarding |
| CLI commands | Every `orc <cmd>` in docs/skills/Makefile exists |
| CLI flags | Flags used in docs match actual `--help` output |
| Docs schema | Directory structure matches docs/README.md spec |
| Getting-started coherence | Guide matches Makefile, bootstrap_cmd.go, first-run skill |
| ER diagram | Tables and relationships match schema.sql |

**Not covered:** Runtime behavior (docs-doctor validates claims, not execution).

## Coverage Matrix

What ORC subsystems are exercised by which integration test:

| Subsystem | self-test | bootstrap-test | bootstrap-exercise | docs-doctor |
|-----------|:---------:|:--------------:|:------------------:|:-----------:|
| Factory CRUD | x | x | x | |
| Workshop CRUD | x | x | x | |
| Workbench CRUD | x | | x | |
| Infra plan/apply | x | | | |
| TMux sessions | x | | | |
| Commission CRUD | | x | x | |
| Shipment lifecycle | | | | |
| Task lifecycle | | | | |
| Note/Tome CRUD | | | | |
| Plan CRUD | | | | |
| Focus system | | | | |
| Status transitions | | | | |
| orc bootstrap cmd | | | x | |
| make bootstrap | | x | | |
| First-run skill | | | x | |
| CLI surface accuracy | | | | x |
| Docs correctness | | | | x |
| ship-* skill chain | | | | |
| deploy workflow | | | | |

## Gaps

Subsystems with no integration-level coverage:

1. **Shipment lifecycle** — No test creates a shipment, transitions it through draft → ready → in-progress → closed, and verifies state at each step.
2. **Task lifecycle** — No test creates tasks, assigns them to shipments, transitions through open → in-progress → closed.
3. **Note/Tome CRUD** — No test creates tomes, attaches notes, lists by container.
4. **Plan CRUD** — No test creates plans attached to tasks.
5. **Focus system** — No test exercises `orc focus`, `orc focus --show`, `orc focus --clear`.
6. **Status transitions** — No test verifies that invalid transitions are rejected (e.g., draft → closed).
7. **Ship skill chain** — The skill sequence (ship-new → ideate → synthesize → plan → run → deploy → complete) is never tested end-to-end.
8. **Deploy workflow** — The merge-to-master flow is never tested (understandably — it's destructive).

## When to Run

| Skill | When |
|-------|------|
| `orc-self-test` | After changes to infra, plan/apply, tmux, or entity CRUD |
| `bootstrap-test` | Before releases, after Makefile/PATH changes |
| `bootstrap-exercise` | After changes to orc-first-run skill or bootstrap command |
| `docs-doctor` | Before merging to master (recommended in CLAUDE.md) |

## Relationship to Unit Tests

These integration skills complement `make test` (the Go test suite). Unit tests cover:
- Guard logic (pure functions)
- Repository correctness (SQL + persistence)
- Service orchestration (mocked ports)
- CLI parsing

Integration skills cover what unit tests can't: real filesystem, real tmux, real CLI invocations, real skill chains, and documentation accuracy. See [testing.md](testing.md) for the unit test pyramid.
