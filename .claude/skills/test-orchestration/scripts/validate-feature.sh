#!/bin/bash
# validate-feature.sh - Validate the POST /echo endpoint implementation
# Usage: ./validate-feature.sh <grove-path>

set -euo pipefail

GROVE_PATH="${1:-}"

if [[ -z "$GROVE_PATH" ]]; then
    echo "status=ERROR"
    echo "error=Missing grove path argument"
    exit 1
fi

if [[ ! -d "$GROVE_PATH" ]]; then
    echo "status=FAIL"
    echo "grove_exists=false"
    echo "error=Grove directory not found: $GROVE_PATH"
    exit 1
fi

echo "grove_exists=true"

cd "$GROVE_PATH"

# Check for required files
echo "=== File Checks ===" >&2

if [[ -f "main.go" ]]; then
    echo "main_go_exists=true"
else
    echo "main_go_exists=false"
    echo "error=main.go not found"
fi

if [[ -f "main_test.go" ]]; then
    echo "main_test_go_exists=true"
else
    echo "main_test_go_exists=false"
    echo "warning=main_test.go not found (tests may be missing)"
fi

if [[ -f "README.md" ]]; then
    echo "readme_exists=true"

    # Check if README mentions /echo endpoint
    if grep -qi "/echo" README.md; then
        echo "readme_has_echo=true"
    else
        echo "readme_has_echo=false"
        echo "warning=/echo endpoint not documented in README"
    fi
else
    echo "readme_exists=false"
    echo "warning=README.md not found"
fi

# Test: go build
echo "=== Build Check ===" >&2

if go build -o /tmp/orc-canary-test 2>&1; then
    echo "build=pass"
    echo "build_exit_code=0"
    rm -f /tmp/orc-canary-test
else
    BUILD_EXIT=$?
    echo "build=fail"
    echo "build_exit_code=$BUILD_EXIT"
    echo "error=go build failed"
fi

# Test: go test
echo "=== Test Check ===" >&2

if [[ -f "main_test.go" ]]; then
    TEST_OUTPUT=$(go test ./... 2>&1)
    TEST_EXIT=$?

    if [[ $TEST_EXIT -eq 0 ]]; then
        echo "tests=pass"
        echo "tests_exit_code=0"

        # Count passed tests
        TESTS_RUN=$(echo "$TEST_OUTPUT" | grep -oE "PASS.*[0-9]+\.[0-9]+s" | wc -l | tr -d ' ')
        echo "tests_run=$TESTS_RUN"
    else
        echo "tests=fail"
        echo "tests_exit_code=$TEST_EXIT"
        echo "error=go test failed"
    fi
else
    echo "tests=skipped"
    echo "tests_exit_code=0"
    echo "warning=No tests found"
fi

# Test: Manual endpoint test (requires starting server)
echo "=== Manual Test Check ===" >&2

# Check if main.go has /echo handler
if grep -q "handleEcho\|/echo" main.go 2>/dev/null; then
    echo "echo_handler_found=true"

    # Start server in background
    go run main.go &
    SERVER_PID=$!
    echo "server_started=true"
    echo "server_pid=$SERVER_PID"

    # Wait for server to start
    sleep 2

    # Try curl test
    CURL_OUTPUT=$(curl -s -X POST http://localhost:8080/echo \
        -H "Content-Type: application/json" \
        -d '{"message":"test"}' 2>&1 || echo "CURL_FAILED")

    if echo "$CURL_OUTPUT" | grep -q "CURL_FAILED"; then
        echo "manual_test=fail"
        echo "error=curl request failed (server may not have started)"
    elif echo "$CURL_OUTPUT" | grep -qi "echo"; then
        echo "manual_test=pass"

        # Validate JSON response
        if echo "$CURL_OUTPUT" | jq -e '.echo' >/dev/null 2>&1; then
            echo "response_valid_json=true"

            # Check if response contains expected fields
            if echo "$CURL_OUTPUT" | jq -e '.echo, .timestamp' >/dev/null 2>&1; then
                echo "response_has_required_fields=true"
            else
                echo "response_has_required_fields=false"
                echo "warning=Response missing echo or timestamp field"
            fi
        else
            echo "response_valid_json=false"
            echo "warning=Response is not valid JSON"
        fi
    else
        echo "manual_test=fail"
        echo "error=Response does not contain 'echo' field"
    fi

    # Kill server
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
    echo "server_stopped=true"
else
    echo "echo_handler_found=false"
    echo "manual_test=skipped"
    echo "error=/echo handler not found in main.go"
fi

# Calculate overall status
echo "=== Overall Status ===" >&2

SUCCESS_COUNT=0
TOTAL_CHECKS=5

# Count successes
[[ -f "main.go" ]] && ((SUCCESS_COUNT++)) || true
[[ "${build:-fail}" == "pass" ]] && ((SUCCESS_COUNT++)) || true
[[ "${tests:-fail}" == "pass" || "${tests:-}" == "skipped" ]] && ((SUCCESS_COUNT++)) || true
[[ "${manual_test:-fail}" == "pass" ]] && ((SUCCESS_COUNT++)) || true
[[ "${readme_has_echo:-false}" == "true" ]] && ((SUCCESS_COUNT++)) || true

echo "success_count=$SUCCESS_COUNT"
echo "total_checks=$TOTAL_CHECKS"

if [[ $SUCCESS_COUNT -ge 4 ]]; then
    echo "status=OK"
elif [[ $SUCCESS_COUNT -ge 3 ]]; then
    echo "status=PARTIAL"
else
    echo "status=FAIL"
fi

exit 0
