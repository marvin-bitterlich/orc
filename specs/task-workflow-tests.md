# Task Workflow FSM - Test Matrix

**Generated:** 2026-01-20
**Source:** `specs/task-workflow.yaml`
**Context:** CON-003 FSM-First Testing Strategy (Slice 3)

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

### 1.1 CanCreateTask

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.1.1 | Mission exists, no shipment | MissionExists: true, ShipmentID: "" | Allowed: true | To implement |
| G1.1.2 | Mission exists, shipment exists | MissionExists: true, ShipmentID: "SHIP-001", ShipmentExists: true | Allowed: true | To implement |
| G1.1.3 | Mission not found | MissionExists: false | Allowed: false, "mission not found" | To implement |
| G1.1.4 | Shipment not found | MissionExists: true, ShipmentID: "SHIP-999", ShipmentExists: false | Allowed: false, "shipment not found" | To implement |

### 1.2 CanCompleteTask

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.2.1 | Unpinned task | IsPinned: false | Allowed: true | To implement |
| G1.2.2 | Pinned task | IsPinned: true | Allowed: false, "cannot complete pinned" | To implement |

### 1.3 CanPauseTask

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.3.1 | In-progress task | Status: "in_progress" | Allowed: true | To implement |
| G1.3.2 | Ready task | Status: "ready" | Allowed: false, "can only pause in_progress" | To implement |
| G1.3.3 | Paused task | Status: "paused" | Allowed: false, "can only pause in_progress" | To implement |
| G1.3.4 | Complete task | Status: "complete" | Allowed: false, "can only pause in_progress" | To implement |

### 1.4 CanResumeTask

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.4.1 | Paused task | Status: "paused" | Allowed: true | To implement |
| G1.4.2 | Ready task | Status: "ready" | Allowed: false, "can only resume paused" | To implement |
| G1.4.3 | In-progress task | Status: "in_progress" | Allowed: false, "can only resume paused" | To implement |
| G1.4.4 | Complete task | Status: "complete" | Allowed: false, "can only resume paused" | To implement |

### 1.5 CanTagTask

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.5.1 | No existing tag | ExistingTagID: "" | Allowed: true | To implement |
| G1.5.2 | Has existing tag | ExistingTagID: "TAG-001", ExistingTagName: "bug" | Allowed: false, "already has tag" | To implement |

### 1.6 GuardResult.Error

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.6.1 | Allowed result | Allowed: true | error == nil | To implement |
| G1.6.2 | Blocked result | Allowed: false, Reason: "test" | error.Error() == "test" | To implement |

---

## 2. Guard Tests Summary

| Guard Function | Test Count | Source Line |
|----------------|------------|-------------|
| CanCreateTask | 4 | task_service.go:31-48 |
| CanCompleteTask | 2 | task_service.go:143-146 |
| CanPauseTask | 4 | task_service.go:158-161 |
| CanResumeTask | 4 | task_service.go:173-176 |
| CanTagTask | 2 | task_service.go:233-241 |
| GuardResult.Error | 2 | (helper method) |
| **Total** | **18** | |

---

## 3. Transition Tests (Service Level)

These tests verify end-to-end transitions through the service layer.

### 3.1 Happy Path Transitions

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T3.1.1 | create | initial | ready | Create task with mission | task_service_test.go | Exists |
| T3.1.2 | claim | ready | in_progress | Claim ready task | task_service_test.go | Exists |
| T3.1.3 | pause | in_progress | paused | Pause in_progress task | task_service_test.go | Exists |
| T3.1.4 | resume | paused | in_progress | Resume paused task | task_service_test.go | Exists |
| T3.1.5 | complete | in_progress | complete | Complete unpinned task | task_service_test.go | Exists |

### 3.2 Self-Transitions

| ID | Transition | State | Test Case | File | Status |
|----|------------|-------|-----------|------|--------|
| T3.2.1 | update | ready | Update title/description | task_service_test.go | Exists |
| T3.2.2 | pin | ready | Pin ready task | task_service_test.go | Exists |
| T3.2.3 | pin | in_progress | Pin in_progress task | task_service_test.go | Exists |
| T3.2.4 | unpin | in_progress | Unpin task | task_service_test.go | Exists |
| T3.2.5 | tag | in_progress | Tag task | task_service_test.go | Exists |
| T3.2.6 | untag | in_progress | Untag task | task_service_test.go | Exists |

---

## 4. Guard Failure Tests (Negative Cases)

### 4.1 mission_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.1.1 | mission_exists | create | Create with missing mission | "mission not found" | guards_test.go |

### 4.2 shipment_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.2.1 | shipment_exists | create | Create with missing shipment | "shipment not found" | guards_test.go |

### 4.3 is_in_progress Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.3.1 | is_in_progress | pause | Pause ready task | "can only pause in_progress" | guards_test.go |
| F4.3.2 | is_in_progress | pause | Pause paused task | "can only pause in_progress" | guards_test.go |
| F4.3.3 | is_in_progress | pause | Pause complete task | "can only pause in_progress" | guards_test.go |

### 4.4 is_paused Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.4.1 | is_paused | resume | Resume ready task | "can only resume paused" | guards_test.go |
| F4.4.2 | is_paused | resume | Resume in_progress task | "can only resume paused" | guards_test.go |
| F4.4.3 | is_paused | resume | Resume complete task | "can only resume paused" | guards_test.go |

### 4.5 not_pinned Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.5.1 | not_pinned | complete | Complete pinned task | "cannot complete pinned" | guards_test.go |

### 4.6 no_existing_tag Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.6.1 | no_existing_tag | tag | Tag task with existing tag | "already has tag" | guards_test.go |

---

## 5. Effect Tests

| ID | Effect | Transition | Test Case | File | Status |
|----|--------|------------|-----------|------|--------|
| E5.1 | WriteDB INSERT | create | Task record created | task_repo_test.go | Exists |
| E5.2 | WriteDB UPDATE | claim | Status + claimed_by set | task_repo_test.go | Exists |
| E5.3 | WriteDB UPDATE | pause | Status updated to paused | task_repo_test.go | Exists |
| E5.4 | WriteDB UPDATE | resume | Status updated to in_progress | task_repo_test.go | Exists |
| E5.5 | WriteDB UPDATE | complete | Status + completed_at | task_repo_test.go | Exists |
| E5.6 | WriteDB INSERT | tag | Task tag created | task_repo_test.go | Exists |
| E5.7 | WriteDB DELETE | untag | Task tag removed | task_repo_test.go | Exists |
| E5.8 | WriteDB DELETE | delete | Task deleted | task_repo_test.go | Exists |

---

## 6. Invariant Tests

| ID | Invariant | Property | Test Strategy | File | Status |
|----|-----------|----------|---------------|------|--------|
| I6.1 | id_format | IDs follow TASK-XXX | Regex check | task_repo_test.go | Exists |
| I6.2 | id_unique | IDs are unique | Duplicate check | task_repo_test.go | Exists |
| I6.3 | status_valid | Status in (ready, in_progress, paused, complete) | Enum check | task_repo_test.go | Exists |
| I6.4 | mission_reference | Mission exists | FK constraint | task_repo_test.go | Exists |
| I6.5 | shipment_reference | Shipment exists (if set) | FK constraint | task_repo_test.go | Exists |
| I6.6 | complete_not_pinned | Complete implies not pinned | Business logic | guards_test.go | New |
| I6.7 | single_tag | Task has at most one tag | Count check | task_repo_test.go | Exists |

---

## 7. Test File Locations

| File | Purpose |
|------|---------|
| `internal/core/task/guards.go` | Guard implementations (NEW) |
| `internal/core/task/guards_test.go` | Guard unit tests (NEW) |
| `internal/app/task_service.go` | Service layer (EXISTS) |
| `internal/app/task_service_test.go` | Service tests (EXISTS) |
| `internal/adapters/sqlite/task_repo.go` | Repository (EXISTS) |
| `internal/adapters/sqlite/task_repo_test.go` | Repository tests (EXISTS) |

---

## 8. Implementation Checklist

### Context Structs to Create

- [ ] `GuardResult` - shared pattern from shipment/grove
- [ ] `CreateTaskContext`
- [ ] `CompleteTaskContext`
- [ ] `StatusTransitionContext`
- [ ] `TagTaskContext`

### Guard Functions to Create

- [ ] `CanCreateTask(ctx CreateTaskContext) GuardResult`
- [ ] `CanCompleteTask(ctx CompleteTaskContext) GuardResult`
- [ ] `CanPauseTask(ctx StatusTransitionContext) GuardResult`
- [ ] `CanResumeTask(ctx StatusTransitionContext) GuardResult`
- [ ] `CanTagTask(ctx TagTaskContext) GuardResult`

### Test Cases to Create (18 total)

**CanCreateTask (4):**
- [ ] TestCanCreateTask_MissionExistsNoShipment
- [ ] TestCanCreateTask_MissionAndShipmentExist
- [ ] TestCanCreateTask_MissionNotFound
- [ ] TestCanCreateTask_ShipmentNotFound

**CanCompleteTask (2):**
- [ ] TestCanCompleteTask_Unpinned
- [ ] TestCanCompleteTask_Pinned

**CanPauseTask (4):**
- [ ] TestCanPauseTask_InProgress
- [ ] TestCanPauseTask_Ready
- [ ] TestCanPauseTask_Paused
- [ ] TestCanPauseTask_Complete

**CanResumeTask (4):**
- [ ] TestCanResumeTask_Paused
- [ ] TestCanResumeTask_Ready
- [ ] TestCanResumeTask_InProgress
- [ ] TestCanResumeTask_Complete

**CanTagTask (2):**
- [ ] TestCanTagTask_NoExistingTag
- [ ] TestCanTagTask_HasExistingTag

**GuardResult.Error (2):**
- [ ] TestGuardResult_Error_Allowed
- [ ] TestGuardResult_Error_Blocked

---

## 9. Comparison: Shipment vs Task Guards

| Aspect | Shipment | Task |
|--------|----------|------|
| Initial state | active | ready |
| Working state | active | in_progress |
| Pause from | active | in_progress |
| Parent guards | mission_exists | mission_exists + shipment_exists |
| Claim operation | No | Yes |
| Tag constraint | No | Yes (one max) |
| Total guards | 5 | 5 |
| Total test cases | 17 | 18 |

---

*Derived from task-workflow.yaml FSM specification*
