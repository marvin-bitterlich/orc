#!/bin/bash
# ORC Integration Test Helpers

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test state
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
TEST_CLEANUP_FUNCS=()

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

log_error() {
    echo -e "${RED}✗ $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_section() {
    echo ""
    echo -e "${BLUE}━━━ $1 ━━━${NC}"
}

# Test assertion functions
assert_command_succeeds() {
    local cmd="$1"
    local description="${2:-command}"

    if eval "$cmd" &>/dev/null; then
        log_success "$description succeeded"
        return 0
    else
        log_error "$description failed"
        return 1
    fi
}

assert_command_fails() {
    local cmd="$1"
    local description="${2:-command}"

    if eval "$cmd" &>/dev/null; then
        log_error "$description succeeded but should have failed"
        return 1
    else
        log_success "$description failed as expected"
        return 0
    fi
}

assert_file_exists() {
    local file="$1"
    local description="${2:-file $file}"

    if [[ -f "$file" ]]; then
        log_success "$description exists"
        return 0
    else
        log_error "$description does not exist"
        return 1
    fi
}

assert_directory_exists() {
    local dir="$1"
    local description="${2:-directory $dir}"

    if [[ -d "$dir" ]]; then
        log_success "$description exists"
        return 0
    else
        log_error "$description does not exist"
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local description="${3:-output}"

    if echo "$haystack" | grep -q "$needle"; then
        log_success "$description contains '$needle'"
        return 0
    else
        log_error "$description does not contain '$needle'"
        return 1
    fi
}

# Test lifecycle functions
register_cleanup() {
    local cleanup_func="$1"
    TEST_CLEANUP_FUNCS+=("$cleanup_func")
}

run_cleanup() {
    log_section "Running cleanup"
    for cleanup_func in "${TEST_CLEANUP_FUNCS[@]}"; do
        log_info "Running: $cleanup_func"
        $cleanup_func || log_warn "Cleanup function $cleanup_func failed (continuing)"
    done
    TEST_CLEANUP_FUNCS=()
}

run_test() {
    local test_name="$1"
    local test_func="$2"

    TESTS_RUN=$((TESTS_RUN + 1))
    log_section "Test: $test_name"

    if $test_func; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        log_success "PASS: $test_name"
        return 0
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        log_error "FAIL: $test_name"
        return 1
    fi
}

print_test_summary() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Test Summary"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Tests Run:    $TESTS_RUN"
    echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}✗ Some tests failed${NC}"
        return 1
    fi
}

# ORC-specific helpers
get_test_mission_id() {
    echo "MISSION-TEST-$(date +%s)"
}

cleanup_test_mission() {
    local mission_id="$1"
    log_info "Cleaning up test mission: $mission_id"

    # TODO: Add proper mission cleanup when implemented
    # For now, just clean up workspace directory
    local mission_dir="$HOME/src/missions/$mission_id"
    if [[ -d "$mission_dir" ]]; then
        rm -rf "$mission_dir"
        log_success "Removed mission directory: $mission_dir"
    fi
}

cleanup_test_grove() {
    local grove_id="$1"
    log_info "Cleaning up test grove: $grove_id"

    # Get grove info
    local grove_path=$(orc grove show "$grove_id" 2>/dev/null | grep "Path:" | awk '{print $2}')

    if [[ -n "$grove_path" ]] && [[ -d "$grove_path" ]]; then
        # Remove worktree
        cd "$HOME/src/orc-canary"
        git worktree remove "$grove_path" --force 2>/dev/null || true
        log_success "Removed worktree: $grove_path"
    fi

    # TODO: Remove grove from database when delete command exists
}

wait_for_tmux_window() {
    local session_name="$1"
    local window_name="$2"
    local max_wait="${3:-5}"

    for i in $(seq 1 $max_wait); do
        if tmux list-windows -t "$session_name" 2>/dev/null | grep -q "$window_name"; then
            return 0
        fi
        sleep 1
    done
    return 1
}
