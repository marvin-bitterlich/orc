#!/bin/bash
# Test: Grove Creation and TMux Integration

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/test-helpers.sh"

# Test identifiers
# MISSION-420 is the dedicated test mission - use it if it exists
TEST_MISSION_ID="MISSION-420"
TEST_MISSION_DIR="$HOME/src/missions/MISSION-420"
TEST_GROVE_NAME="test-canary-$(date +%s)"
# NOTE: Grove path includes mission ID prefix (e.g., MISSION-XXX-grove-name)
TEST_GROVE_PATH=""  # Will be set after mission is known
CANARY_REPO="$HOME/src/orc-canary"
CREATED_MISSION=false

# Cleanup function
cleanup() {
    log_section "Cleanup"

    # Remove worktree if it exists
    if [[ -n "$TEST_GROVE_PATH" ]] && [[ -d "$TEST_GROVE_PATH" ]]; then
        cd "$CANARY_REPO"
        git worktree remove "$TEST_GROVE_PATH" --force 2>/dev/null || true
        log_success "Removed worktree"
    fi

    # Only remove mission directory if we created it (not MISSION-420)
    if [[ "$CREATED_MISSION" == "true" ]] && [[ -d "$TEST_MISSION_DIR" ]]; then
        rm -rf "$TEST_MISSION_DIR"
        log_success "Removed test mission directory"
    fi

    # Kill test TMux session if it exists
    if tmux has-session -t "test-orc-session" 2>/dev/null; then
        tmux kill-session -t "test-orc-session"
        log_success "Killed test TMux session"
    fi
}

trap cleanup EXIT

# Setup: Use MISSION-420 (dedicated test mission) or create one
setup_mission() {
    log_info "Setting up test mission"

    # Check if MISSION-420 exists
    if orc mission list 2>&1 | grep -q "MISSION-420"; then
        log_info "Using dedicated test mission: MISSION-420"
        TEST_MISSION_ID="MISSION-420"
        TEST_MISSION_DIR="$HOME/src/missions/MISSION-420"
    else
        log_warn "MISSION-420 not found, creating temporary mission"

        local output=$(orc mission create "Grove Test Mission" \
            --description "Testing grove and TMux integration" 2>&1)

        # Extract mission ID from output
        TEST_MISSION_ID=$(echo "$output" | grep -oE 'MISSION-[0-9]+' | head -1)

        if [[ -z "$TEST_MISSION_ID" ]]; then
            log_error "Failed to extract mission ID from output"
            return 1
        fi

        log_info "Created temporary mission: $TEST_MISSION_ID"
        TEST_MISSION_DIR="$HOME/src/missions/$TEST_MISSION_ID"
        CREATED_MISSION=true
    fi

    mkdir -p "$TEST_MISSION_DIR/.orc"

    cat > "$TEST_MISSION_DIR/.orc-mission" <<EOF
{
  "mission_id": "$TEST_MISSION_ID",
  "workspace_path": "$TEST_MISSION_DIR",
  "created_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF

    cat > "$TEST_MISSION_DIR/.orc/config.json" <<EOF
{
  "version": "1.0",
  "type": "mission",
  "mission": {
    "mission_id": "$TEST_MISSION_ID",
    "current_epic": ""
  }
}
EOF

    # Set grove path now that we know the mission ID
    TEST_GROVE_PATH="$HOME/src/worktrees/$TEST_MISSION_ID-$TEST_GROVE_NAME"

    log_success "Test mission ready: $TEST_MISSION_ID"
}

# Test 1: Create grove with worktree
test_create_grove() {
    log_info "Creating grove with git worktree"

    cd "$TEST_MISSION_DIR"

    # Create grove
    orc grove create "$TEST_GROVE_NAME" \
        --repos orc-canary \
        --mission "$TEST_MISSION_ID"

    # Verify grove in database
    assert_command_succeeds "orc grove list | grep -q \"$TEST_GROVE_NAME\"" \
        "Grove appears in grove list"

    # Verify worktree exists
    assert_directory_exists "$TEST_GROVE_PATH" "Worktree directory"

    # Verify it's a valid git directory
    assert_command_succeeds "cd '$TEST_GROVE_PATH' && git status" \
        "Worktree is valid git directory"

    return $?
}

# Test 2: Verify grove metadata
test_grove_metadata() {
    log_info "Verifying grove metadata"

    assert_directory_exists "$TEST_GROVE_PATH/.orc" "Grove .orc directory"
    # Grove now uses config.json instead of metadata.json
    assert_file_exists "$TEST_GROVE_PATH/.orc/config.json" "Grove config.json"

    local config=$(cat "$TEST_GROVE_PATH/.orc/config.json")
    assert_contains "$config" "$TEST_MISSION_ID" "config.json contains mission ID"

    return $?
}

# Test 3: Verify mission context propagation via config.json
test_mission_context_in_grove() {
    log_info "Verifying mission context in grove config.json"

    # Grove stores mission context in .orc/config.json (not separate .orc-mission file)
    assert_file_exists "$TEST_GROVE_PATH/.orc/config.json" "Grove config.json exists"

    local config=$(cat "$TEST_GROVE_PATH/.orc/config.json")
    assert_contains "$config" "$TEST_MISSION_ID" "config.json contains mission ID"
    assert_contains "$config" "mission_id" "config.json has mission_id field"

    return $?
}

# Test 4: Test orc commands from grove directory
test_commands_from_grove() {
    log_info "Testing orc commands from grove directory"

    cd "$TEST_GROVE_PATH"

    # Status should show correct mission
    local status_output=$(orc status 2>&1)
    assert_contains "$status_output" "$TEST_MISSION_ID" "orc status shows correct mission from grove"

    # Summary should work
    assert_command_succeeds "orc summary" "orc summary works from grove directory"

    return $?
}

# Test 5: TMux integration (basic - no actual pane testing in automated env)
test_tmux_session_basics() {
    log_info "Testing TMux session basics"

    # Create test TMux session
    if ! tmux has-session -t "test-orc-session" 2>/dev/null; then
        tmux new-session -d -s "test-orc-session" -c "$TEST_MISSION_DIR"
        log_success "Created test TMux session"
    fi

    # Verify session exists
    assert_command_succeeds "tmux has-session -t 'test-orc-session'" \
        "TMux session exists"

    # Verify we can list windows
    assert_command_succeeds "tmux list-windows -t 'test-orc-session'" \
        "Can list TMux windows"

    return $?
}

# Test 6: Verify grove show command
test_grove_show() {
    log_info "Testing grove show command"

    cd "$TEST_MISSION_DIR"

    # Get grove ID
    local grove_id=$(orc grove list | grep "$TEST_GROVE_NAME" | awk '{print $1}')

    if [[ -z "$grove_id" ]]; then
        log_error "Could not find grove ID"
        return 1
    fi

    # Show grove details
    local grove_details=$(orc grove show "$grove_id" 2>&1)
    assert_contains "$grove_details" "$TEST_GROVE_NAME" "Grove show displays name"
    assert_contains "$grove_details" "$TEST_MISSION_ID" "Grove show displays mission"
    assert_contains "$grove_details" "$TEST_GROVE_PATH" "Grove show displays path"

    return $?
}

# Run all tests
main() {
    log_section "Grove Creation and TMux Integration Tests"

    # Setup
    setup_mission

    # Run tests
    run_test "Create grove with worktree" test_create_grove
    run_test "Verify grove metadata" test_grove_metadata
    run_test "Verify mission context in grove" test_mission_context_in_grove
    run_test "Test commands from grove directory" test_commands_from_grove
    run_test "TMux session basics" test_tmux_session_basics
    run_test "Grove show command" test_grove_show

    print_test_summary
}

main
