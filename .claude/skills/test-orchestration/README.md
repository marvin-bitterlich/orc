# Orchestration Test Skill

Full real-world integration test of the ORC orchestration system using Claude Code skills.

## What This Does

This skill performs the **ultimate validation** of ORC by:

1. ✓ Creating a real test mission
2. ✓ Spinning up full TMux session (deputy ORC + IMPs)
3. ✓ Giving them REAL work (add a feature to canary app)
4. ✓ Monitoring their progress through checkpoints
5. ✓ Verifying they complete the work correctly
6. ✓ Generating comprehensive report with all validation results

**If this works, we prove ORC can orchestrate complex development tasks autonomously.**

## Quick Start

### Prerequisites

- ORC installed and in PATH
- ORC initialized (`orc init`)
- orc-canary repository at `~/src/orc-canary`
- TMux installed
- jq installed (for JSON validation)

### Run the Test

```bash
cd ~/src/orc
claude --skill test-orchestration
```

Or invoke via Claude Code CLI:

```bash
/test-orchestration
```

The skill will execute automatically, running through all 7 phases and generating progress reports in the `turns/` directory.

### Check Results

After execution completes, check:

```bash
# Human-readable final report
cat ~/src/orc/skills/test-orchestration/turns/06-final-report.md

# Machine-readable results
cat ~/src/orc/skills/test-orchestration/turns/results.json

# Individual phase reports
ls ~/src/orc/skills/test-orchestration/turns/
```

## Skill Structure

```
skills/test-orchestration/
├── README.md              # This file
├── manifest.json          # Skill metadata
├── skill.md               # Main skill prompt (execution logic)
├── config.json            # Test configuration
├── TEST-WORKLOAD.md       # Feature specification for IMPs
├── scripts/               # Helper utilities
│   ├── verify-tmux-layout.sh      # Verify TMux session structure
│   ├── check-deputy-health.sh     # Verify deputy ORC operational
│   ├── monitor-imp-progress.sh    # Watch IMP activity
│   ├── validate-feature.sh        # Test implemented feature
│   └── cleanup-test-env.sh        # Clean test environment
└── turns/                 # Progress output (created at runtime)
    ├── 00-setup.md
    ├── 01-grove-deployed.md
    ├── 02-deputy-verified.md
    ├── 03-work-assigned.md
    ├── 04-progress-N.md
    ├── 05-validation.md
    ├── 06-final-report.md
    └── results.json
```

## Configuration

Edit `config.json` to customize test parameters:

- **Mission settings**: Title template, workspace path
- **Grove settings**: Repository, worktree base path
- **TMux settings**: Session prefix, window names
- **Test workload**: Feature to implement (POST /echo endpoint)
- **Validation**: Checkpoints, required files, test commands
- **Monitoring**: Progress check interval, timeout
- **Cleanup**: What to clean up on success/failure

## Execution Phases

The skill executes in 7 sequential phases:

### Phase 1: Environment Setup
- Create test mission with auto-generated ID
- Provision workspace directory structure
- Write `.orc-mission` and metadata files
- **Checkpoints**: 4

### Phase 2: Deploy TMux Session
- Create grove from orc-canary repository
- Launch TMux session with deputy and IMP windows
- Set up 3-pane IMP layout (vim | claude | shell)
- **Checkpoints**: 5

### Phase 3: Verify Deputy ORC
- Run health checks on deputy ORC
- Verify mission context detection
- Test `orc status` and `orc summary` commands
- **Checkpoints**: 4

### Phase 4: Assign Real Work
- Create parent work order: "Implement POST /echo endpoint"
- Create 4 child work orders for specific tasks
- Verify all work orders visible in deputy summary
- **Checkpoints**: 3

### Phase 5: Monitor Implementation
- Watch IMPs work on the feature
- Track file changes in grove
- Monitor work order status updates
- Write progress reports periodically
- **Checkpoints**: 4

### Phase 6: Validate Results
- Test code compiles (`go build`)
- Test all tests pass (`go test`)
- Manual curl test of endpoint
- Verify README updated
- **Checkpoints**: 5

### Phase 7: Generate Report & Cleanup
- Compile all phase results
- Calculate success rate (25 total checkpoints)
- Write final report and results.json
- Clean up test environment
- **Exit code**: 0 if success, 1 if failure

**Total Checkpoints**: 25 across all phases

## Test Workload

The skill creates these work orders for IMPs to implement:

**Feature**: POST /echo endpoint

**Work Orders**:
1. Add POST /echo handler to main.go
2. Write unit tests for /echo endpoint
3. Update README with /echo documentation
4. Run tests and verify implementation

**Success Criteria**:
- Code compiles
- Tests pass
- Manual test works: `curl -X POST http://localhost:8080/echo -d '{"message":"test"}'`
- README updated

See `TEST-WORKLOAD.md` for complete specification.

## Helper Scripts

All scripts output key=value format for easy parsing:

### verify-tmux-layout.sh
```bash
./scripts/verify-tmux-layout.sh orc-MISSION-TEST-XXX
# Output: status=OK, session_exists=true, panes=4, layout=valid
```

### check-deputy-health.sh
```bash
./scripts/check-deputy-health.sh MISSION-TEST-XXX
# Output: status=OK, context=deputy, mission=MISSION-TEST-XXX
```

### monitor-imp-progress.sh
```bash
./scripts/monitor-imp-progress.sh orc-MISSION-TEST-XXX MISSION-TEST-XXX
# Output: work_orders=4, completed=2, in_progress=1, progress_percent=50
```

### validate-feature.sh
```bash
./scripts/validate-feature.sh ~/src/worktrees/test-canary-XXX
# Output: status=OK, build=pass, tests=pass, manual_test=pass
```

### cleanup-test-env.sh
```bash
./scripts/cleanup-test-env.sh MISSION-TEST-XXX
# Output: status=OK, mission_deleted=true, grove_removed=true
```

## Success Criteria

**Test PASSES if**:
- All 25 checkpoints pass
- POST /echo endpoint implemented correctly
- All tests pass
- Manual endpoint test works
- README documentation updated

**Test FAILS if**:
- Any critical phase fails (setup, deployment, health check, work assignment)
- Feature validation has less than 4/5 checkpoints passing
- Cleanup fails

## Output Files

### turns/06-final-report.md

Human-readable final report with:
- Executive summary (PASS/FAIL, duration, success rate)
- Phase-by-phase results (all 6 phases with checkpoints)
- Feature validation details
- Performance metrics
- Recommendations

### turns/results.json

Machine-readable results for CI/CD integration:

```json
{
  "test_run_id": "test-20260114-123456",
  "start_time": "2026-01-14T12:34:56Z",
  "end_time": "2026-01-14T12:45:23Z",
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

## Troubleshooting

### "orc command not found"
```bash
cd ~/src/orc
go build -o orc cmd/orc/main.go
# Add to PATH or use absolute path
```

### "orc-canary repository not found"
```bash
cd ~/src
git clone git@github.com:example/orc-canary.git
```

### "TMux session already exists"
```bash
# Kill existing test session
tmux kill-session -t orc-MISSION-TEST-XXX
```

### Test gets stuck in Phase 5 (Implementation)
- Check IMP pane is active: `tmux attach -t orc-MISSION-TEST-XXX`
- Verify work orders are visible: `orc summary` from mission workspace
- Check for errors in IMP shell pane
- The test will timeout after 30 minutes and proceed to validation

### Cleanup fails
```bash
# Manual cleanup
./scripts/cleanup-test-env.sh MISSION-TEST-XXX

# Or manually:
rm -rf ~/src/missions/MISSION-TEST-*
rm -rf ~/src/worktrees/test-canary-*
cd ~/src/orc-canary && git worktree prune
tmux kill-session -t orc-MISSION-TEST-*
```

## CI/CD Integration

This skill is designed for CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run ORC Orchestration Test
  run: |
    cd ~/src/orc
    claude --skill test-orchestration

    # Check exit code
    if [ $? -eq 0 ]; then
      echo "✓ Orchestration test passed"
    else
      echo "✗ Orchestration test failed"
      cat skills/test-orchestration/turns/06-final-report.md
      exit 1
    fi
```

## Performance

Expected execution time:
- **Fast path** (everything works): 15-20 minutes
- **Typical** (some debugging needed): 20-25 minutes
- **Maximum** (timeout): 30 minutes

The 30-minute timeout is generous to account for:
- Claude thinking time
- File I/O operations
- TMux session startup
- IMPs reading context and planning
- Implementation work
- Test execution

## Why This Matters

If this test passes, it proves that ORC can:

✓ Orchestrate complex development tasks autonomously
✓ Coordinate multiple Claude agents (master, deputy, IMPs)
✓ Handle real-world workflows (mission → grove → work orders → implementation)
✓ Maintain context across agent boundaries
✓ Deliver working features end-to-end

**Success = We're unstoppable** 🚀

## Development

### Adding New Validation Checks

1. Edit `skill.md` to add checkpoint in appropriate phase
2. Update `config.json` checkpoint counts
3. Update helper scripts if needed
4. Update this README

### Changing Test Workload

1. Edit `TEST-WORKLOAD.md` with new feature spec
2. Update `config.json` test_workload section
3. Update `validate-feature.sh` to test new feature
4. Update skill.md Phase 4 and Phase 6

### Adding New Helper Scripts

1. Create script in `scripts/` directory
2. Make it executable: `chmod +x scripts/new-script.sh`
3. Output key=value format for easy parsing
4. Update skill.md to reference it
5. Update this README

## License

Part of the ORC project. See main repository for license details.

---

**Status**: ✓ Ready for testing
**Maintained by**: ORC Development Team
**Last Updated**: 2026-01-14
