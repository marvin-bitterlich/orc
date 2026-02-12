# Testing & Verification

## Table-Driven Tests (Default Pattern)

Default to table-driven tests for guards, planners, validation, and service decision logic.

```go
func TestCanCloseTask(t *testing.T) {
    tests := []struct {
        name        string
        ctx         StatusTransitionContext
        wantAllowed bool
        wantReason  string
    }{
        {
            name: "can close in_progress task",
            ctx:  StatusTransitionContext{TaskID: "TASK-001", Status: "in_progress"},
            wantAllowed: true,
        },
        {
            name: "cannot close open task",
            ctx:  StatusTransitionContext{TaskID: "TASK-001", Status: "open"},
            wantAllowed: false,
            wantReason:  "can only close in_progress tasks (current status: open)",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CanCloseTask(tt.ctx)
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

## Test Pyramid

```
+-----------------------------+
|     Integration Tests       |  <- Sparse: end-to-end CLI flows / wiring
+-----------------------------+
|     Repository Tests        |  <- Medium: SQL correctness + persistence invariants
+-----------------------------+
|      Service Tests          |  <- Most: orchestration logic (mock ports)
+-----------------------------+
|       Guard Tests           |  <- Foundation: pure functions
+-----------------------------+
```

## Test Helpers

Use `testutil_test.go` helpers in `internal/adapters/sqlite/` where available to avoid repeating DB setup + seeding.

## Test Commission for CLI Validation

When developing changes that affect CLI display (summary, containers, leafs, etc.), use the test commission to validate output:

```bash
./orc summary --commission COMM-003
```

### What's in COMM-003

The test commission contains representative examples of all container types:
- **Shipment** (SHIP-205) with tasks and notes
- **Tome** (TOME-008, standalone) with notes

Also includes items with various statuses (draft, ready, in-progress, closed, open) and pinned items.

### When to Use

Run `orc summary --commission COMM-003` after changes to:
- Summary display logic
- Container creation/update commands
- Leaf item (task, note, plan) display
- Status filtering or colorization
- Pinned item display
- Hierarchical nesting

### Maintenance

If you add new container types or display features, add corresponding test data to COMM-003.

## Verification Discipline

LLMs are prone to skipping checks. ORC's workflow requires explicit verification.

### Plans Must Include Checks

Every implementation plan must explicitly list:
- [ ] Tests to run
- [ ] Lint checks to pass
- [ ] Manual verification steps (if applicable)

### Completion Must Report What Ran (and what didn't)

When completing work, report verification explicitly:

```
Ran: make test (all passing)
Ran: make lint (no issues)
Skipped: <check> (reason)
```

**Rule:** If a check was not run, it must be explicitly marked as skipped with a reason. Never imply success.

## Bootstrap VM Testing

See [integration-tests.md](integration-tests.md) for full details on the bootstrap VM test and all other integration test skills.
