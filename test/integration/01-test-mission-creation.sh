#!/bin/bash
# Test: Mission Creation and Deputy Bootstrap

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/test-helpers.sh"

# Test mission ID (will be set after creation)
TEST_MISSION_ID=""
TEST_MISSION_DIR=""

# Cleanup function
cleanup() {
    log_section "Cleanup"
    if [[ -n "$TEST_MISSION_ID" ]]; then
        # Delete from database (function added for testing)
        # For now, just clean up directory
        log_info "Cleaning up mission: $TEST_MISSION_ID"
    fi
    if [[ -n "$TEST_MISSION_DIR" ]] && [[ -d "$TEST_MISSION_DIR" ]]; then
        rm -rf "$TEST_MISSION_DIR"
        log_success "Removed test mission directory"
    fi
}

# Register cleanup to run on exit
trap cleanup EXIT

# Test 1: Create mission
test_create_mission() {
    log_info "Creating test mission"

    local output=$(orc mission create "Integration Test Mission" \
        --description "Automated integration test mission" 2>&1)

    # Extract mission ID from output (format: "✓ Created mission MISSION-XXX: ...")
    TEST_MISSION_ID=$(echo "$output" | grep -oE 'MISSION-[0-9]+' | head -1)

    if [[ -z "$TEST_MISSION_ID" ]]; then
        log_error "Failed to extract mission ID from output"
        return 1
    fi

    log_info "Created mission: $TEST_MISSION_ID"
    TEST_MISSION_DIR="$HOME/src/missions/$TEST_MISSION_ID"

    assert_command_succeeds "orc mission list | grep -q '$TEST_MISSION_ID'" \
        "Mission appears in mission list"

    return $?
}

# Test 2: Create mission workspace directory
test_create_mission_workspace() {
    log_info "Creating mission workspace directory"

    mkdir -p "$TEST_MISSION_DIR/.orc"
    assert_directory_exists "$TEST_MISSION_DIR" "Mission workspace directory"
    assert_directory_exists "$TEST_MISSION_DIR/.orc" "Mission .orc directory"

    return $?
}

# Test 3: Write .orc-mission marker
test_write_mission_marker() {
    log_info "Writing .orc-mission marker"

    cat > "$TEST_MISSION_DIR/.orc-mission" <<EOF
{
  "mission_id": "$TEST_MISSION_ID",
  "workspace_path": "$TEST_MISSION_DIR",
  "created_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF

    assert_file_exists "$TEST_MISSION_DIR/.orc-mission" ".orc-mission marker file"

    return $?
}

# Test 4: Write workspace metadata.json
test_write_workspace_metadata() {
    log_info "Writing workspace metadata.json"

    cat > "$TEST_MISSION_DIR/.orc/metadata.json" <<EOF
{
  "active_mission_id": "$TEST_MISSION_ID",
  "last_updated": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF

    assert_file_exists "$TEST_MISSION_DIR/.orc/metadata.json" "Workspace metadata.json"

    return $?
}

# Test 5: Verify mission context detection
test_mission_context_detection() {
    log_info "Testing mission context detection"

    cd "$TEST_MISSION_DIR"

    # Run orc status and check output
    local status_output=$(orc status 2>&1)

    assert_contains "$status_output" "$TEST_MISSION_ID" "orc status shows test mission"
    assert_contains "$status_output" "Deputy Context" "orc status detects deputy context"

    return $?
}

# Test 6: Create test work order
test_create_work_order() {
    log_info "Creating test work order"

    cd "$TEST_MISSION_DIR"

    # Create work order (should auto-scope to this mission)
    orc work-order create "Test work order for integration testing"

    # Verify it appears in summary
    local summary_output=$(orc summary 2>&1)
    assert_contains "$summary_output" "$TEST_MISSION_ID" "orc summary shows test mission"
    assert_contains "$summary_output" "Test work order" "Summary shows test work order"

    return $?
}

# Test 7: Verify command auto-scoping
test_command_autoscoping() {
    log_info "Testing command auto-scoping in deputy context"

    cd "$TEST_MISSION_DIR"

    # Get work order list - should only show this mission's orders
    local wo_list=$(orc work-order list 2>&1)

    # Should contain our test mission
    assert_contains "$wo_list" "$TEST_MISSION_ID" "Work order list scoped to test mission"

    # Should NOT contain other missions (like MISSION-001)
    if echo "$wo_list" | grep -q "MISSION-001"; then
        log_error "Work order list includes other missions (should be scoped)"
        return 1
    else
        log_success "Work order list properly scoped to test mission only"
    fi

    return 0
}

# Run all tests
main() {
    log_section "Mission Creation and Deputy Bootstrap Tests"

    run_test "Create mission" test_create_mission
    run_test "Create mission workspace" test_create_mission_workspace
    run_test "Write .orc-mission marker" test_write_mission_marker
    run_test "Write workspace metadata" test_write_workspace_metadata
    run_test "Mission context detection" test_mission_context_detection
    run_test "Create work order in deputy context" test_create_work_order
    run_test "Command auto-scoping" test_command_autoscoping

    print_test_summary
}

main
