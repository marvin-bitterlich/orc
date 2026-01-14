---
name: test-orchestration
description: Full real-world orchestration test that creates a test mission, spins up TMux session with deputy + IMPs, assigns real work, monitors progress, validates completion, and generates comprehensive report. Use this to validate the entire ORC multi-agent coordination workflow.
---

# Orchestration Test: Full Real-World Multi-Agent Validation

You are executing a comprehensive integration test of the ORC orchestration system. This test validates the ENTIRE multi-agent coordination workflow by creating a real mission, spinning up deputy ORC + IMPs in TMux, assigning them actual development work, and verifying they complete it correctly.

## Your Mission

Execute a 7-phase orchestration test that proves ORC can autonomously coordinate multiple Claude agents to complete real development tasks. This is the ultimate validation of the system.

## Critical Rules

1. **Execute ALL phases sequentially** - Do not skip any phase
2. **Validate checkpoints** - Each phase has specific validation criteria that MUST pass
3. **Write progress to turns/** - Document every phase in markdown files
4. **Use helper scripts** - They're in scripts/ directory for verification tasks
5. **Handle errors gracefully** - If a checkpoint fails, document it and decide whether to continue or abort
6. **Generate final report** - turns/07-final-report.md with complete results

## Test Configuration

Load configuration from `config.json` in this skill directory. Key parameters:
- Mission workspace: `~/src/missions/`
- Grove worktree: `~/src/worktrees/`
- Canary repo: `~/src/orc-canary`
- Test workload: POST /echo endpoint implementation
- Timeout: 30 minutes max

## Phase 1: Environment Setup

**Goal**: Create test mission and provision workspace

### Tasks

1. Generate unique mission ID: `MISSION-TEST-ORC-{timestamp}`
2. Create mission using ORC CLI:
   ```bash
   orc mission create "Orchestration Test Mission" \
     --description "Automated orchestration test - validates multi-agent coordination"
   ```
3. Create mission workspace directory structure:
   ```bash
   mkdir -p ~/src/missions/{MISSION_ID}/.orc
   ```
4. Write `.orc-mission` marker (JSON format):
   ```json
   {
     "mission_id": "MISSION-TEST-ORC-...",
     "workspace_path": "~/src/missions/MISSION-TEST-ORC-...",
     "created_at": "2026-01-14T12:34:56Z"
   }
   ```
5. Write `.orc/metadata.json`:
   ```json
   {
     "active_mission_id": "MISSION-TEST-ORC-...",
     "last_updated": "2026-01-14T12:34:56Z"
   }
   ```

### Validation Checkpoints (4 total)

- [ ] Mission created with correct ID format (MISSION-TEST-ORC-{timestamp})
- [ ] Mission workspace directory exists at `~/src/missions/{MISSION_ID}`
- [ ] `.orc-mission` marker file contains valid JSON with mission_id
- [ ] `.orc/metadata.json` file contains active_mission_id

### Output

Write `turns/00-setup.md` with:
- Mission ID
- Workspace path
- All created files
- Validation results (✓ or ✗ for each checkpoint)
- Status: PASS or FAIL

**If any checkpoint fails, ABORT and write final report**

## Phase 2: Deploy TMux Session

**Goal**: Create grove from orc-canary and launch full TMux environment

### Tasks

1. Create grove with git worktree:
   ```bash
   cd ~/src/missions/{MISSION_ID}
   orc grove create test-canary-{timestamp} \
     --repos orc-canary \
     --mission {MISSION_ID}
   ```
2. Verify grove created and worktree exists
3. Launch TMux session:
   ```bash
   tmux new-session -d -s orc-{MISSION_ID} -c ~/src/missions/{MISSION_ID}
   ```
4. Create deputy window (pane 0):
   ```bash
   tmux rename-window -t orc-{MISSION_ID}:0 deputy
   tmux send-keys -t orc-{MISSION_ID}:deputy "cd ~/src/missions/{MISSION_ID}" C-m
   tmux send-keys -t orc-{MISSION_ID}:deputy "claude" C-m
   ```
5. Create IMP window with 3-pane layout:
   ```bash
   cd ~/src/worktrees/test-canary-{timestamp}
   orc grove open {GROVE_ID}
   ```
   (This creates new window with vim | claude | shell layout)

### Validation Checkpoints (5 total)

- [ ] Grove exists in database (`orc grove list`)
- [ ] Worktree directory exists at `~/src/worktrees/test-canary-{timestamp}`
- [ ] TMux session exists (`tmux has-session -t orc-{MISSION_ID}`)
- [ ] Deputy window exists (`tmux list-windows -t orc-{MISSION_ID} | grep deputy`)
- [ ] IMP window exists with correct layout (use `scripts/verify-tmux-layout.sh`)

### Output

Write `turns/01-grove-deployed.md` with:
- Grove ID and path
- TMux session name
- Window/pane layout details
- Validation results
- Status: PASS or FAIL

**If any checkpoint fails, run cleanup and ABORT**

## Phase 3: Verify Deputy ORC

**Goal**: Ensure deputy ORC is operational and detects mission context correctly

### Tasks

1. Run health check script:
   ```bash
   ./scripts/check-deputy-health.sh {MISSION_ID}
   ```
2. Verify context detection by checking deputy pane output
3. Test `orc status` shows correct mission
4. Test `orc summary` scopes to mission only

### Validation Checkpoints (4 total)

- [ ] Deputy context detected (check-deputy-health.sh returns OK)
- [ ] `orc status` shows test mission ID
- [ ] `orc summary` displays deputy context header
- [ ] Can create work orders in deputy context

### Output

Write `turns/02-deputy-verified.md` with:
- Health check results
- Context detection confirmation
- Sample orc status/summary output
- Validation results
- Status: PASS or FAIL

**If any checkpoint fails, run cleanup and ABORT**

## Phase 4: Assign Real Work

**Goal**: Create work orders for POST /echo endpoint implementation

### Tasks

1. Create parent work order:
   ```bash
   cd ~/src/missions/{MISSION_ID}
   orc work-order create "Implement POST /echo endpoint" \
     --description "Add echo endpoint to canary app with tests and documentation"
   ```
2. Create 4 child work orders (from config.json test_workload.child_wos):
   - WO-A: Add POST /echo handler to main.go
   - WO-B: Write unit tests for /echo endpoint
   - WO-C: Update README with /echo endpoint documentation
   - WO-D: Run tests and verify implementation

3. Verify all work orders visible in deputy summary

### Validation Checkpoints (3 total)

- [ ] Parent work order created successfully
- [ ] All 4 child work orders created
- [ ] `orc summary` shows all work orders scoped to test mission

### Output

Write `turns/03-work-assigned.md` with:
- Parent work order ID and title
- All child work order IDs and titles
- orc summary output showing work orders
- Validation results
- Status: PASS or FAIL

**If any checkpoint fails, run cleanup and ABORT**

## Phase 5: Monitor Implementation

**Goal**: Watch IMPs work on the feature and track progress

**NOTE**: This phase is OBSERVATIONAL. You are NOT implementing the feature yourself. You are monitoring the IMP Claude instances working in the TMux panes.

### Tasks

1. Start monitoring script:
   ```bash
   ./scripts/monitor-imp-progress.sh orc-{MISSION_ID}
   ```
2. Check for file changes in grove:
   ```bash
   cd ~/src/worktrees/test-canary-{timestamp}
   git status
   ```
3. Periodically check work order status:
   ```bash
   cd ~/src/missions/{MISSION_ID}
   orc work-order list
   ```
4. Write progress updates to `turns/04-progress-N.md` every 2-3 minutes
5. Wait until either:
   - All work orders marked complete, OR
   - Timeout reached (30 minutes), OR
   - Implementation appears stuck (no changes for 10 minutes)

### Validation Checkpoints (4 total)

- [ ] Files modified in grove (main.go, main_test.go, README.md exist)
- [ ] Git shows uncommitted changes
- [ ] At least some work orders marked complete
- [ ] No errors visible in IMP pane

### Output

Write `turns/04-progress-N.md` (multiple files) with:
- Timestamp
- Files changed
- Work order status updates
- IMP activity observations
- Current state: in_progress, completed, or stuck

**If timeout reached or stuck, proceed to validation anyway to see what was completed**

## Phase 6: Validate Results

**Goal**: Test the implemented feature and verify it works correctly

### Tasks

1. Run validation script:
   ```bash
   ./scripts/validate-feature.sh ~/src/worktrees/test-canary-{timestamp}
   ```
2. Manual checks:
   - Code compiles: `cd {grove_path} && go build`
   - Tests pass: `go test ./...`
   - Feature works: Start server, run `curl -X POST http://localhost:8080/echo -d '{"message":"test"}'`
   - README updated: Check for /echo documentation
3. Check work order completion status
4. Review git changes: `git diff`

### Validation Checkpoints (5 total)

- [ ] `go build` succeeds (exit code 0)
- [ ] `go test ./...` passes (exit code 0)
- [ ] Manual curl test returns correct JSON response
- [ ] README.md contains /echo endpoint documentation
- [ ] Feature meets all requirements from work orders

### Output

Write `turns/05-validation.md` with:
- Build results
- Test results
- Manual test results (curl output)
- README verification
- Work order completion status
- Validation results (how many of 5 checkpoints passed)
- Status: PASS or FAIL

## Phase 7: Generate Report & Cleanup

**Goal**: Create comprehensive final report and clean up test environment

### Tasks

1. Compile all phase results
2. Calculate overall success rate (checkpoints passed / total checkpoints)
3. Write `turns/06-final-report.md` with:
   - Executive summary (pass/fail, duration, success rate)
   - Phase-by-phase results (all 6 phases)
   - Feature validation details
   - Performance metrics (time to provision, time to implement)
   - Recommendations
4. Write `turns/results.json` (machine-readable results):
   ```json
   {
     "test_run_id": "test-{timestamp}",
     "start_time": "...",
     "end_time": "...",
     "duration_seconds": 627,
     "overall_result": "PASS",
     "phases": {
       "setup": {"checkpoints_passed": 4, "checkpoints_total": 4, "status": "PASS"},
       "tmux_deploy": {"checkpoints_passed": 5, "checkpoints_total": 5, "status": "PASS"},
       "deputy_health": {"checkpoints_passed": 4, "checkpoints_total": 4, "status": "PASS"},
       "work_assigned": {"checkpoints_passed": 3, "checkpoints_total": 3, "status": "PASS"},
       "implementation": {"checkpoints_passed": 4, "checkpoints_total": 4, "status": "PASS"},
       "validation": {"checkpoints_passed": 5, "checkpoints_total": 5, "status": "PASS"}
     },
     "success_rate": 1.0
   }
   ```
5. Run cleanup script:
   ```bash
   ./scripts/cleanup-test-env.sh {MISSION_ID}
   ```
   (Only if config.cleanup_on_success is true AND test passed)

### Output

Write `turns/06-final-report.md` and `turns/results.json`

Exit with code 0 if all 25 checkpoints passed, otherwise exit with code 1

## Helper Scripts Reference

Located in `scripts/` directory:

1. **verify-tmux-layout.sh** - Verifies TMux session structure
   ```bash
   ./scripts/verify-tmux-layout.sh orc-{MISSION_ID}
   # Output: status=OK, session_exists=true, panes=4, layout=valid
   ```

2. **check-deputy-health.sh** - Verifies deputy ORC operational
   ```bash
   ./scripts/check-deputy-health.sh {MISSION_ID}
   # Output: status=OK, context=deputy, mission={MISSION_ID}
   ```

3. **monitor-imp-progress.sh** - Watches IMP activity
   ```bash
   ./scripts/monitor-imp-progress.sh orc-{MISSION_ID}
   # Output: work_orders=4, completed=2, in_progress=1, blocked=0
   ```

4. **validate-feature.sh** - Tests implemented feature
   ```bash
   ./scripts/validate-feature.sh {grove_path}
   # Output: status=OK, build=pass, tests=pass, manual_test=pass
   ```

5. **cleanup-test-env.sh** - Cleans up test environment
   ```bash
   ./scripts/cleanup-test-env.sh {MISSION_ID}
   # Output: status=OK, mission_deleted=true, grove_removed=true
   ```

## Success Criteria

**Overall test PASSES if**:
- All 25 checkpoints pass (4+5+4+3+4+5)
- POST /echo endpoint implemented correctly
- Tests pass
- Manual test works
- Environment cleans up properly

**Overall test FAILS if**:
- Any critical phase fails (setup, tmux deploy, deputy health, work assignment)
- Feature validation fails (less than 4/5 checkpoints)
- Cleanup fails

## Error Handling

If any phase fails:
1. Document the failure in the phase's turn file
2. Decide: Can we continue or must we abort?
3. If aborting: Skip to Phase 7 (cleanup and final report)
4. If continuing: Note the failure and proceed

## Time Management

- **Total budget**: 30 minutes
- **Phase 1-4**: Should complete in <5 minutes
- **Phase 5**: Up to 20 minutes for implementation
- **Phase 6-7**: <5 minutes

If you exceed 30 minutes total, proceed to validation and cleanup anyway.

## Final Note

This is the ultimate test of ORC's orchestration capabilities. If this succeeds, it proves ORC can autonomously coordinate multi-agent development workflows end-to-end.

**Execute with precision. Document everything. Generate the comprehensive report.**

¡Vamos! 🚀
