# Conclave Workflow FSM - Test Matrix

**Generated:** 2026-01-20
**Source:** `specs/conclave-workflow.yaml`
**Context:** CON-003 FSM-First Testing Strategy (Slice 4)

---

## Test Categories

| Category | Description | Priority |
|----------|-------------|----------|
| Guard Tests | Pure guard function evaluation | P0 |
| Transition Tests | Happy path state transitions | P1 |
| Guard Failure Tests | Negative cases where guards block | P1 |
| Effect Tests | Side effects execute correctly | P2 |
| Invariant Tests | Property-based constraints | P3 |

---

## 1. Guard Tests (Pure Functions)

### 1.1 CanCreateConclave

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.1.1 | Mission exists | MissionExists: true | Allowed: true | To implement |
| G1.1.2 | Mission not found | MissionExists: false | Allowed: false, "mission not found" | To implement |

### 1.2 CanCompleteConclave

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.2.1 | Unpinned conclave | IsPinned: false | Allowed: true | To implement |
| G1.2.2 | Pinned conclave | IsPinned: true | Allowed: false, "cannot complete pinned" | To implement |

### 1.3 CanPauseConclave

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.3.1 | Active conclave | Status: "active" | Allowed: true | To implement |
| G1.3.2 | Paused conclave | Status: "paused" | Allowed: false, "can only pause active" | To implement |
| G1.3.3 | Complete conclave | Status: "complete" | Allowed: false, "can only pause active" | To implement |

### 1.4 CanResumeConclave

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.4.1 | Paused conclave | Status: "paused" | Allowed: true | To implement |
| G1.4.2 | Active conclave | Status: "active" | Allowed: false, "can only resume paused" | To implement |
| G1.4.3 | Complete conclave | Status: "complete" | Allowed: false, "can only resume paused" | To implement |

### 1.5 GuardResult.Error

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.5.1 | Allowed result | Allowed: true | error == nil | To implement |
| G1.5.2 | Blocked result | Allowed: false, Reason: "test" | error.Error() == "test" | To implement |

---

## 2. Guard Tests Summary

| Guard Function | Test Count | Source Line |
|----------------|------------|-------------|
| CanCreateConclave | 2 | conclave_service.go:28-34 |
| CanCompleteConclave | 2 | conclave_service.go:100-103 |
| CanPauseConclave | 3 | conclave_service.go:115-118 |
| CanResumeConclave | 3 | conclave_service.go:130-133 |
| GuardResult.Error | 2 | (helper method) |
| **Total** | **12** | |

---

## 3. Transition Tests (Service Level)

These tests verify end-to-end transitions through the service layer.

### 3.1 Happy Path Transitions

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T3.1.1 | create | initial | active | Create conclave with mission | conclave_service_test.go | Exists |
| T3.1.2 | pause | active | paused | Pause active conclave | conclave_service_test.go | Exists |
| T3.1.3 | resume | paused | active | Resume paused conclave | conclave_service_test.go | Exists |
| T3.1.4 | complete | active | complete | Complete unpinned conclave | conclave_service_test.go | Exists |

### 3.2 Self-Transitions

| ID | Transition | State | Test Case | File | Status |
|----|------------|-------|-----------|------|--------|
| T3.2.1 | update | active | Update title/description | conclave_service_test.go | Exists |
| T3.2.2 | pin | active | Pin conclave | conclave_service_test.go | Exists |
| T3.2.3 | unpin | active | Unpin conclave | conclave_service_test.go | Exists |

---

## 4. Guard Failure Tests (Negative Cases)

### 4.1 mission_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.1.1 | mission_exists | create | Create with missing mission | "mission not found" | guards_test.go |

### 4.2 is_active Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.2.1 | is_active | pause | Pause paused conclave | "can only pause active" | guards_test.go |
| F4.2.2 | is_active | pause | Pause complete conclave | "can only pause active" | guards_test.go |

### 4.3 is_paused Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.3.1 | is_paused | resume | Resume active conclave | "can only resume paused" | guards_test.go |
| F4.3.2 | is_paused | resume | Resume complete conclave | "can only resume paused" | guards_test.go |

### 4.4 not_pinned Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.4.1 | not_pinned | complete | Complete pinned conclave | "cannot complete pinned" | guards_test.go |

---

## 5. Effect Tests

| ID | Effect | Transition | Test Case | File | Status |
|----|--------|------------|-----------|------|--------|
| E5.1 | WriteDB INSERT | create | Conclave record created | conclave_repo_test.go | Exists |
| E5.2 | WriteDB UPDATE | pause | Status updated to paused | conclave_repo_test.go | Exists |
| E5.3 | WriteDB UPDATE | resume | Status updated to active | conclave_repo_test.go | Exists |
| E5.4 | WriteDB UPDATE | complete | Status + completed_at | conclave_repo_test.go | Exists |
| E5.5 | WriteDB DELETE | delete | Conclave deleted | conclave_repo_test.go | Exists |

---

## 6. Invariant Tests

| ID | Invariant | Property | Test Strategy | File | Status |
|----|-----------|----------|---------------|------|--------|
| I6.1 | id_format | IDs follow CON-XXX | Regex check | conclave_repo_test.go | Exists |
| I6.2 | id_unique | IDs are unique | Duplicate check | conclave_repo_test.go | Exists |
| I6.3 | status_valid | Status in (active, paused, complete) | Enum check | conclave_repo_test.go | Exists |
| I6.4 | mission_reference | Mission exists | FK constraint | conclave_repo_test.go | Exists |
| I6.5 | complete_not_pinned | Complete implies not pinned | Business logic | guards_test.go | New |

---

## 7. Test File Locations

| File | Purpose |
|------|---------|
| `internal/core/conclave/guards.go` | Guard implementations (NEW) |
| `internal/core/conclave/guards_test.go` | Guard unit tests (NEW) |
| `internal/app/conclave_service.go` | Service layer (EXISTS) |
| `internal/app/conclave_service_test.go` | Service tests (EXISTS) |
| `internal/adapters/sqlite/conclave_repo.go` | Repository (EXISTS) |
| `internal/adapters/sqlite/conclave_repo_test.go` | Repository tests (EXISTS) |

---

## 8. Implementation Checklist

### Context Structs to Create

- [x] `GuardResult` - shared pattern from shipment/grove packages
- [ ] `CreateConclaveContext`
- [ ] `CompleteConclaveContext`
- [ ] `StatusTransitionContext`

### Guard Functions to Create

- [ ] `CanCreateConclave(ctx CreateConclaveContext) GuardResult`
- [ ] `CanCompleteConclave(ctx CompleteConclaveContext) GuardResult`
- [ ] `CanPauseConclave(ctx StatusTransitionContext) GuardResult`
- [ ] `CanResumeConclave(ctx StatusTransitionContext) GuardResult`

### Test Cases to Create (12 total)

**CanCreateConclave (2):**
- [ ] TestCanCreateConclave_MissionExists
- [ ] TestCanCreateConclave_MissionNotFound

**CanCompleteConclave (2):**
- [ ] TestCanCompleteConclave_Unpinned
- [ ] TestCanCompleteConclave_Pinned

**CanPauseConclave (3):**
- [ ] TestCanPauseConclave_Active
- [ ] TestCanPauseConclave_Paused
- [ ] TestCanPauseConclave_Complete

**CanResumeConclave (3):**
- [ ] TestCanResumeConclave_Paused
- [ ] TestCanResumeConclave_Active
- [ ] TestCanResumeConclave_Complete

**GuardResult.Error (2):**
- [ ] TestGuardResult_Error_Allowed
- [ ] TestGuardResult_Error_Blocked

---

## 9. Comparison: Shipment vs Conclave Guards

| Aspect | Shipment | Conclave |
|--------|----------|----------|
| ORC-only guards | No | No |
| Status guards | is_active, is_paused, not_pinned | is_active, is_paused, not_pinned |
| Grove exclusivity | grove_not_assigned_elsewhere | None |
| Total guards | 5 | 4 |
| Total test cases | 16 | 12 |

---

*Derived from conclave-workflow.yaml FSM specification*
