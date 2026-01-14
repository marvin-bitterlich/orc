# ORC Integration Tests

Comprehensive integration test suite for validating ORC core functionality with real-world scenarios.

## Quick Start

```bash
# Run all integration tests
cd ~/src/orc
./test/integration/run-all-tests.sh

# Run individual test
./test/integration/01-test-mission-creation.sh
./test/integration/02-test-grove-tmux.sh
```

## Test Coverage

### 01-test-mission-creation.sh (7 tests)
Tests mission creation and deputy ORC bootstrap workflow:
- ✓ Create mission
- ✓ Create mission workspace directory structure
- ✓ Write .orc-mission marker (JSON format)
- ✓ Write workspace metadata.json
- ✓ Mission context detection from .orc-mission file
- ✓ Create work order in deputy context (auto-scoping)
- ✓ Command auto-scoping to mission

**Validates:**
- Mission lifecycle
- Deputy context detection
- File-based mission markers
- Auto-scoping of commands to mission

### 02-test-grove-tmux.sh (6 tests)
Tests grove creation and TMux integration:
- ✓ Create grove with git worktree
- ✓ Verify grove metadata in .orc/ directory
- ✓ Verify mission marker propagation to grove
- ✓ Test commands from grove directory
- ✓ TMux session basics
- ✓ Grove show command

**Validates:**
- Git worktree integration
- Grove database registration
- Metadata propagation
- Cross-directory command execution
- TMux session management

## Test Framework

### test-helpers.sh
Provides reusable test utilities:

**Assertion Functions:**
- `assert_command_succeeds` - Test command success
- `assert_command_fails` - Test expected failure
- `assert_file_exists` - Verify file existence
- `assert_directory_exists` - Verify directory existence
- `assert_contains` - Check string containment

**Test Lifecycle:**
- `run_test` - Execute test with pass/fail tracking
- `register_cleanup` - Register cleanup functions
- `run_cleanup` - Execute all cleanup functions
- `print_test_summary` - Display test results

**Logging:**
- `log_info` - Informational messages (blue)
- `log_success` - Success messages (green)
- `log_error` - Error messages (red)
- `log_warn` - Warning messages (yellow)
- `log_section` - Section headers

### Test Isolation

Each test:
- Creates unique missions/groves (using timestamps)
- Registers cleanup functions
- Cleans up on exit (success or failure)
- Does not interfere with other tests

## Prerequisites

- ORC installed and in PATH (`cd ~/src/orc && go build && install orc`)
- ORC initialized (`orc init` creates ~/.orc/orc.db)
- TMux installed
- orc-canary repository at ~/src/orc-canary
- Git configured

## Running Tests

### Full Test Suite
```bash
cd ~/src/orc
./test/integration/run-all-tests.sh
```

**Output:**
```
━━━ ORC Integration Test Suite ━━━
Running: 01-test-mission-creation.sh
  ✓ 7/7 tests passed
Running: 02-test-grove-tmux.sh
  ✓ 6/6 tests passed

✓✓✓ ALL TESTS PASSED ✓✓✓
Test Files Run: 2
Test Files Passed: 2
Test Files Failed: 0
```

### Individual Tests
```bash
# Mission creation tests only
./test/integration/01-test-mission-creation.sh

# Grove and TMux tests only
./test/integration/02-test-grove-tmux.sh
```

## Test Development

### Adding New Tests

1. Create test file: `test/integration/XX-test-name.sh`
2. Source helpers: `source "$SCRIPT_DIR/test-helpers.sh"`
3. Define test functions: `test_something() { ... }`
4. Register cleanup: `cleanup() { ... }; trap cleanup EXIT`
5. Run tests: `run_test "Description" test_something`
6. Add to `run-all-tests.sh` TEST_FILES array

### Example Test Function
```bash
test_something() {
    log_info "Testing something"

    # Do test actions
    orc command ...

    # Make assertions
    assert_command_succeeds "orc status" "Status command works"
    assert_contains "$output" "expected" "Output contains text"

    return $?
}
```

### Cleanup Pattern
```bash
TEST_MISSION_ID=""

cleanup() {
    log_section "Cleanup"
    if [[ -n "$TEST_MISSION_ID" ]]; then
        # Clean up test data
        rm -rf "$HOME/src/missions/$TEST_MISSION_ID"
    fi
}

trap cleanup EXIT
```

## Test Scenarios Validated

### Mission Lifecycle
- ✓ Mission creation with auto-generated IDs
- ✓ Mission workspace directory structure
- ✓ .orc-mission marker (JSON format with mission_id)
- ✓ Workspace metadata.json (active_mission_id)
- ✓ Mission context detection from any subdirectory

### Deputy ORC
- ✓ Context auto-detection from .orc-mission file
- ✓ Command auto-scoping to mission (no --mission flag needed)
- ✓ Work order creation scoped to mission
- ✓ Summary and status commands show correct mission

### Grove Management
- ✓ Grove creation with database registration
- ✓ Git worktree creation and validation
- ✓ Metadata propagation (.orc/metadata.json)
- ✓ Mission marker propagation (.orc-mission)
- ✓ Commands work from grove directories
- ✓ Grove show command displays correct info

### TMux Integration
- ✓ TMux session creation
- ✓ Window/pane management
- ✓ Directory context preservation

## Known Limitations

1. **TMux Pane Testing**: Tests verify TMux sessions and windows exist but don't test actual pane content or IMP spawning (requires interactive environment)

2. **Mission Deletion**: Currently only cleans up directories, not database records (DeleteMission function exists but not exposed via CLI yet)

3. **Parallel Execution**: Tests should be run sequentially to avoid TMux session conflicts

## CI/CD Integration

These tests are designed to run in CI environments:

```bash
# In CI pipeline
- name: Run ORC Integration Tests
  run: |
    cd ~/src/orc
    ./test/integration/run-all-tests.sh
```

**Exit Codes:**
- 0 = All tests passed
- 1 = Some tests failed

## Troubleshooting

### "orc command not found"
```bash
cd ~/src/orc
go build -o orc cmd/orc/main.go
# Add to PATH or use absolute path
```

### "ORC database not found"
```bash
orc init
```

### "orc-canary repository not found"
```bash
cd ~/src
git clone git@github.com:looneym/orc-canary.git
```

### "TMux session already exists"
```bash
# Kill existing test session
tmux kill-session -t test-orc-session
```

### Stale test missions/groves
```bash
# Manual cleanup
rm -rf ~/src/missions/MISSION-*
rm -rf ~/src/worktrees/test-canary-*
cd ~/src/orc-canary && git worktree prune
```

## Test Results

Last run: 2026-01-14

```
Test Files: 2
Total Tests: 13
Passed: 13 ✓
Failed: 0
Success Rate: 100%
```

## Roadmap

Future test additions:
- [ ] Work order state transitions
- [ ] Handoff creation and retrieval
- [ ] Grove open command (IMP layout validation)
- [ ] Proto-mail system (WO-061 ↔ WO-065)
- [ ] Cross-grove coordination
- [ ] Error cases and edge conditions
- [ ] Performance benchmarks

## Contributing

When adding new ORC features:
1. Write integration tests first (TDD)
2. Ensure tests are isolated and clean up properly
3. Update this README with new test coverage
4. Run full test suite before committing

---

**Status**: ✓ Operational - All core functionality validated
**Maintained by**: ORC Development Team
**Last Updated**: 2026-01-14
