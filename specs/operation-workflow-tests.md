# Operation Workflow FSM - Test Matrix

**Generated:** 2026-01-20
**Source:** `specs/operation-workflow.yaml`
**Context:** CON-003 FSM-First Testing Strategy (Slice 8)

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

### 1.1 CanCreateOperation

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.1.1 | Mission exists | MissionExists: true | Allowed: true | To implement |
| G1.1.2 | Mission not found | MissionExists: false | Allowed: false, "mission not found" | To implement |

### 1.2 CanCompleteOperation

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.2.1 | Ready status | Status: "ready" | Allowed: true | To implement |
| G1.2.2 | Already complete | Status: "complete" | Allowed: false, "can only complete ready" | To implement |

### 1.3 GuardResult.Error

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.3.1 | Allowed result | Allowed: true | error == nil | To implement |
| G1.3.2 | Blocked result | Allowed: false, Reason: "test" | error.Error() == "test" | To implement |

---

## 2. Guard Tests Summary

| Guard Function | Test Count | Source Line |
|----------------|------------|-------------|
| CanCreateOperation | 2 | internal/core/operation/guards.go |
| CanCompleteOperation | 2 | internal/core/operation/guards.go |
| GuardResult.Error | 2 | (helper method) |
| **Total** | **6** | |

---

## 3. Transition Tests (Service Level)

These tests verify end-to-end transitions through the service layer.

### 3.1 Happy Path Transitions

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T3.1.1 | create | initial | ready | Create operation with mission | operation_service_test.go | Exists |
| T3.1.2 | complete | ready | complete | Complete ready operation | operation_service_test.go | Exists |

### 3.2 Self-Transitions

| ID | Transition | State | Test Case | File | Status |
|----|------------|-------|-----------|------|--------|
| T3.2.1 | update | ready | Update title/description | operation_service_test.go | Exists |

---

## 4. Guard Failure Tests (Negative Cases)

### 4.1 mission_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.1.1 | mission_exists | create | Create with missing mission | "mission not found" | guards_test.go |

### 4.2 is_ready Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.2.1 | is_ready | complete | Complete already completed operation | "can only complete ready" | guards_test.go |

---

## 5. Effect Tests

| ID | Effect | Transition | Test Case | File | Status |
|----|--------|------------|-----------|------|--------|
| E5.1 | WriteDB INSERT | create | Operation record created | operation_repo_test.go | Exists |
| E5.2 | WriteDB UPDATE | update | Title/description updated | operation_repo_test.go | Exists |
| E5.3 | WriteDB UPDATE | complete | Status + completed_at | operation_repo_test.go | Exists |

---

## 6. Invariant Tests

| ID | Invariant | Property | Test Strategy | File | Status |
|----|-----------|----------|---------------|------|--------|
| I6.1 | id_format | IDs follow OP-XXX | Regex check | operation_repo_test.go | Exists |
| I6.2 | id_unique | IDs are unique | Duplicate check | operation_repo_test.go | Exists |
| I6.3 | status_valid | Status in (ready, complete) | Enum check | operation_repo_test.go | Exists |
| I6.4 | mission_reference | Mission exists | FK constraint | operation_repo_test.go | Exists |
| I6.5 | complete_requires_completed_at | Complete implies completed_at | Business logic | guards_test.go | New |

---

## 7. Test File Locations

| File | Purpose |
|------|---------|
| `internal/core/operation/guards.go` | Guard implementations (NEW) |
| `internal/core/operation/guards_test.go` | Guard unit tests (NEW) |
| `internal/app/operation_service.go` | Service layer (EXISTS) |
| `internal/app/operation_service_test.go` | Service tests (EXISTS) |
| `internal/adapters/sqlite/operation_repo.go` | Repository (EXISTS) |
| `internal/adapters/sqlite/operation_repo_test.go` | Repository tests (EXISTS) |

---

## 8. Implementation Checklist

### Context Structs to Create

- [x] `GuardResult` - shared pattern from other packages
- [ ] `CreateOperationContext`
- [ ] `CompleteOperationContext`

### Guard Functions to Create

- [ ] `CanCreateOperation(ctx CreateOperationContext) GuardResult`
- [ ] `CanCompleteOperation(ctx CompleteOperationContext) GuardResult`

### Test Cases to Create (6 total)

**CanCreateOperation (2):**
- [ ] TestCanCreateOperation_MissionExists
- [ ] TestCanCreateOperation_MissionNotFound

**CanCompleteOperation (2):**
- [ ] TestCanCompleteOperation_ReadyStatus
- [ ] TestCanCompleteOperation_AlreadyComplete

**GuardResult.Error (2):**
- [ ] TestGuardResult_Error_Allowed
- [ ] TestGuardResult_Error_Blocked

---

## 9. Comparison: Operation vs Other Entities

| Aspect | Operation | Plan | Question | Task |
|--------|-----------|------|----------|------|
| States | 2 (ready, complete) | 2 (draft, approved) | 2 (open, answered) | 4 (active, paused, etc.) |
| Terminal states | 1 (complete) | 2 (approved, deleted) | 2 (answered, deleted) | 2 (complete, deleted) |
| Pin/Unpin | No | Yes | Yes | Yes |
| Delete | No | Yes | Yes | Yes |
| Pause/Resume | No | No | No | Yes |
| Optional parent | None | Shipment | Investigation | Shipment |
| Guards | 2 | 3 | 2 | 4 |
| Test cases | 6 | 14 | 10 | 14 |
| ID format | OP-XXX | PLAN-XXX | Q-XXX | TASK-XXX |

---

## 10. Key Simplifications

Operation is intentionally the simplest entity in ORC:

1. **No pin/unpin:** No need to protect operations from accidental changes
2. **No delete:** Operations serve as permanent audit trail
3. **No pause/resume:** Linear workflow from ready to complete
4. **Single parent:** Only Mission (no optional secondary parent)
5. **Minimal guards:** Only validate mission exists and ready status

This simplicity makes Operation ideal for tracking discrete, atomic units of work.

---

*Derived from operation-workflow.yaml FSM specification*
