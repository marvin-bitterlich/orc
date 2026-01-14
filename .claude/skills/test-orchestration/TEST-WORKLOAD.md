# Test Workload: POST /echo Endpoint

This document defines the test workload that IMPs will implement during orchestration testing.

## Feature Overview

**Endpoint**: `POST /echo`

**Purpose**: Simple echo service that accepts a message and returns it with a timestamp.

**Why This Works as a Test**:
- Simple enough to implement quickly (~10-20 minutes)
- Complex enough to require actual work (structs, handler, tests, docs)
- Clear validation criteria (curl test, unit tests)
- Requires multiple file changes (main.go, main_test.go, README.md)
- Tests full development workflow (code → test → document → verify)

## Requirements Specification

### Endpoint Behavior

**Request**:
```http
POST /echo HTTP/1.1
Content-Type: application/json

{
  "message": "Hello, ORC!"
}
```

**Response** (200 OK):
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "echo": "Hello, ORC!",
  "timestamp": "2026-01-14T12:34:56Z"
}
```

**Error Response** (400 Bad Request):
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "Invalid request"
}
```

### Implementation Requirements

#### 1. Add Structs to main.go

```go
type EchoRequest struct {
    Message string `json:"message"`
}

type EchoResponse struct {
    Echo      string `json:"echo"`
    Timestamp string `json:"timestamp"`
}
```

#### 2. Add Handler Function to main.go

```go
func handleEcho(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req EchoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    if req.Message == "" {
        http.Error(w, "Message cannot be empty", http.StatusBadRequest)
        return
    }

    resp := EchoResponse{
        Echo:      req.Message,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

#### 3. Register Route in main()

```go
func main() {
    // ... existing code ...

    http.HandleFunc("/echo", handleEcho)

    // ... rest of main ...
}
```

#### 4. Create Unit Tests in main_test.go

Required tests:
- Test valid request returns correct response
- Test invalid JSON returns 400
- Test empty message returns 400
- Test method other than POST returns 405

Example test structure:
```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHandleEcho_ValidRequest(t *testing.T) {
    reqBody := EchoRequest{Message: "test message"}
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/echo", bytes.NewReader(body))
    w := httptest.NewRecorder()

    handleEcho(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }

    var resp EchoResponse
    json.NewDecoder(w.Body).Decode(&resp)

    if resp.Echo != "test message" {
        t.Errorf("Expected echo 'test message', got '%s'", resp.Echo)
    }

    if resp.Timestamp == "" {
        t.Error("Expected timestamp to be set")
    }
}

// ... more tests ...
```

#### 5. Update README.md

Add to endpoints section:

```markdown
### POST /echo

Echo service that returns the provided message with a timestamp.

**Request:**
```json
{
  "message": "Hello, ORC!"
}
```

**Response:**
```json
{
  "echo": "Hello, ORC!",
  "timestamp": "2026-01-14T12:34:56Z"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/echo \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello, ORC!"}'
```
```

## Work Order Breakdown

The test creates these work orders for IMPs to complete:

### Parent Work Order

**Title**: Implement POST /echo endpoint

**Description**: Add echo endpoint to canary app with tests and documentation

**Acceptance Criteria**:
- POST /echo endpoint implemented
- Unit tests written and passing
- README updated with endpoint documentation
- Manual test works: `curl -X POST http://localhost:8080/echo -d '{"message":"test"}'`

### Child Work Orders

#### WO-A: Add POST /echo handler to main.go

**Tasks**:
- Create EchoRequest struct with message field
- Create EchoResponse struct with echo and timestamp fields
- Implement handleEcho function
- Register route: `http.HandleFunc("/echo", handleEcho)`
- Handle method validation (POST only)
- Handle JSON decoding errors
- Handle empty message validation

**Files Modified**: `main.go`

#### WO-B: Write unit tests

**Tasks**:
- Create main_test.go
- Test valid request returns correct response
- Test invalid JSON returns 400 error
- Test empty message returns 400 error
- Test non-POST method returns 405 error
- Ensure all tests pass

**Files Created**: `main_test.go`

#### WO-C: Update README

**Tasks**:
- Add POST /echo to endpoint list
- Document request format
- Document response format
- Add curl example
- Add example request/response

**Files Modified**: `README.md`

#### WO-D: Run tests and verify

**Tasks**:
- Run `go test ./...` and ensure all pass
- Run `go build` and ensure no errors
- Start server and run manual curl test
- Verify response format is correct
- Commit changes

**Validation**: All tests pass, build succeeds, manual test works

## Success Criteria

The feature is considered successfully implemented when:

1. ✓ Code compiles without errors (`go build`)
2. ✓ All unit tests pass (`go test ./...`)
3. ✓ Manual curl test returns correct JSON response
4. ✓ README.md contains /echo documentation
5. ✓ Response includes both "echo" and "timestamp" fields
6. ✓ Timestamp is in RFC3339 format
7. ✓ Empty message returns 400 error
8. ✓ Invalid JSON returns 400 error

## Validation Commands

These commands will be run during Phase 6 validation:

```bash
# 1. Build check
cd {grove_path}
go build
# Expected: exit code 0

# 2. Test check
go test ./...
# Expected: exit code 0, all tests pass

# 3. Manual test
go run main.go &
sleep 2
curl -X POST http://localhost:8080/echo \
  -H "Content-Type: application/json" \
  -d '{"message":"test"}'
# Expected: {"echo":"test","timestamp":"2026-01-14T..."}

# 4. README check
grep -i "/echo" README.md
# Expected: match found

# 5. Response validation
# Check JSON is valid and has required fields
curl ... | jq -e '.echo, .timestamp'
# Expected: both fields present
```

## Expected Timeline

- **WO-A** (handler implementation): 3-5 minutes
- **WO-B** (unit tests): 5-7 minutes
- **WO-C** (README update): 2-3 minutes
- **WO-D** (verification): 2-3 minutes

**Total**: 12-18 minutes for experienced Claude developer

**Test timeout**: 30 minutes (generous buffer)

## Notes

This workload is intentionally:
- **Self-contained**: No external dependencies
- **Testable**: Clear pass/fail criteria
- **Realistic**: Actual feature development work
- **Achievable**: Can be completed in reasonable time
- **Verifiable**: Automated validation possible

It serves as the ultimate test of ORC's ability to coordinate IMPs on real development tasks.
