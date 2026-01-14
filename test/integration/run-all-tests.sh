#!/bin/bash
# ORC Integration Test Runner
# Runs all integration tests and reports results

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/test-helpers.sh"

# Test files to run (in order)
TEST_FILES=(
    "01-test-mission-creation.sh"
    "02-test-grove-tmux.sh"
)

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Run all test files
run_all_tests() {
    log_section "ORC Integration Test Suite"
    log_info "Running all integration tests..."
    echo ""

    for test_file in "${TEST_FILES[@]}"; do
        local test_path="$SCRIPT_DIR/$test_file"

        if [[ ! -f "$test_path" ]]; then
            log_warn "Test file not found: $test_file (skipping)"
            continue
        fi

        log_section "Running: $test_file"

        if bash "$test_path"; then
            PASSED_TESTS=$((PASSED_TESTS + 1))
            log_success "✓ $test_file PASSED"
        else
            FAILED_TESTS=$((FAILED_TESTS + 1))
            log_error "✗ $test_file FAILED"
        fi

        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        echo ""
    done
}

# Print final summary
print_final_summary() {
    echo ""
    echo "════════════════════════════════════════════"
    echo "  ORC Integration Test Suite Summary"
    echo "════════════════════════════════════════════"
    echo ""
    echo "Test Files Run:    $TOTAL_TESTS"
    echo -e "Test Files Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Test Files Failed: ${RED}$FAILED_TESTS${NC}"
    echo ""

    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo -e "${GREEN}✓✓✓ ALL TESTS PASSED ✓✓✓${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}✗✗✗ SOME TESTS FAILED ✗✗✗${NC}"
        echo ""
        return 1
    fi
}

# Main execution
main() {
    # Ensure we're in the right place
    if [[ ! -f "$HOME/.orc/orc.db" ]]; then
        log_error "ORC database not found. Have you run 'orc init'?"
        exit 1
    fi

    # Check for required dependencies
    if ! command -v orc &>/dev/null; then
        log_error "orc command not found in PATH"
        log_info "Please build and install orc: cd ~/src/orc && go build && install orc"
        exit 1
    fi

    if ! command -v tmux &>/dev/null; then
        log_error "tmux not found in PATH"
        exit 1
    fi

    if [[ ! -d "$HOME/src/orc-canary" ]]; then
        log_error "orc-canary repository not found at ~/src/orc-canary"
        exit 1
    fi

    # Run tests
    run_all_tests

    # Print summary
    print_final_summary
}

main "$@"
