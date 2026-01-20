# Shipment Workflow FSM - Test Matrix

**Generated:** 2026-01-20
**Source:** `specs/shipment-workflow.yaml`
**Context:** CON-003 FSM-First Testing Strategy (Slice 2)

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

### 1.1 CanCreateShipment

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.1.1 | Mission exists | MissionExists: true | Allowed: true | To implement |
| G1.1.2 | Mission not found | MissionExists: false | Allowed: false, "mission not found" | To implement |

### 1.2 CanCompleteShipment

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.2.1 | Unpinned shipment | IsPinned: false | Allowed: true | To implement |
| G1.2.2 | Pinned shipment | IsPinned: true | Allowed: false, "cannot complete pinned" | To implement |

### 1.3 CanPauseShipment

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.3.1 | Active shipment | Status: "active" | Allowed: true | To implement |
| G1.3.2 | Paused shipment | Status: "paused" | Allowed: false, "can only pause active" | To implement |
| G1.3.3 | Complete shipment | Status: "complete" | Allowed: false, "can only pause active" | To implement |

### 1.4 CanResumeShipment

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.4.1 | Paused shipment | Status: "paused" | Allowed: true | To implement |
| G1.4.2 | Active shipment | Status: "active" | Allowed: false, "can only resume paused" | To implement |
| G1.4.3 | Complete shipment | Status: "complete" | Allowed: false, "can only resume paused" | To implement |

### 1.5 CanAssignGrove

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.5.1 | Unassigned grove | ShipmentExists: true, GroveAssignedToID: "" | Allowed: true | To implement |
| G1.5.2 | Grove assigned to same shipment | ShipmentExists: true, GroveAssignedToID: "SHIP-001" (same) | Allowed: true | To implement |
| G1.5.3 | Grove assigned to other shipment | ShipmentExists: true, GroveAssignedToID: "SHIP-002" | Allowed: false, "grove already assigned" | To implement |
| G1.5.4 | Shipment not found | ShipmentExists: false | Allowed: false, "shipment not found" | To implement |

### 1.6 GuardResult.Error

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.6.1 | Allowed result | Allowed: true | error == nil | To implement |
| G1.6.2 | Blocked result | Allowed: false, Reason: "test" | error.Error() == "test" | To implement |

---

## 2. Guard Tests Summary

| Guard Function | Test Count | Source Line |
|----------------|------------|-------------|
| CanCreateShipment | 2 | shipment_service.go:32-38 |
| CanCompleteShipment | 2 | shipment_service.go:104-107 |
| CanPauseShipment | 3 | shipment_service.go:119-122 |
| CanResumeShipment | 3 | shipment_service.go:134-137 |
| CanAssignGrove | 4 | shipment_service.go:170-177 |
| GuardResult.Error | 2 | (helper method) |
| **Total** | **16+** | |

---

## 3. Transition Tests (Service Level)

These tests verify end-to-end transitions through the service layer.

### 3.1 Happy Path Transitions

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T3.1.1 | create | initial | active | Create shipment with mission | shipment_service_test.go | Exists |
| T3.1.2 | pause | active | paused | Pause active shipment | shipment_service_test.go | Exists |
| T3.1.3 | resume | paused | active | Resume paused shipment | shipment_service_test.go | Exists |
| T3.1.4 | complete | active | complete | Complete unpinned shipment | shipment_service_test.go | Exists |

### 3.2 Self-Transitions

| ID | Transition | State | Test Case | File | Status |
|----|------------|-------|-----------|------|--------|
| T3.2.1 | update | active | Update title/description | shipment_service_test.go | Exists |
| T3.2.2 | pin | active | Pin shipment | shipment_service_test.go | Exists |
| T3.2.3 | unpin | active | Unpin shipment | shipment_service_test.go | Exists |
| T3.2.4 | assign_grove | active | Assign grove | shipment_service_test.go | Exists |

---

## 4. Guard Failure Tests (Negative Cases)

### 4.1 mission_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.1.1 | mission_exists | create | Create with missing mission | "mission not found" | guards_test.go |

### 4.2 is_active Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.2.1 | is_active | pause | Pause paused shipment | "can only pause active" | guards_test.go |
| F4.2.2 | is_active | pause | Pause complete shipment | "can only pause active" | guards_test.go |

### 4.3 is_paused Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.3.1 | is_paused | resume | Resume active shipment | "can only resume paused" | guards_test.go |
| F4.3.2 | is_paused | resume | Resume complete shipment | "can only resume paused" | guards_test.go |

### 4.4 not_pinned Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.4.1 | not_pinned | complete | Complete pinned shipment | "cannot complete pinned" | guards_test.go |

### 4.5 grove_not_assigned_elsewhere Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.5.1 | grove_not_assigned | assign_grove | Grove assigned elsewhere | "grove already assigned" | guards_test.go |
| F4.5.2 | shipment_exists | assign_grove | Shipment not found | "shipment not found" | guards_test.go |

---

## 5. Effect Tests

| ID | Effect | Transition | Test Case | File | Status |
|----|--------|------------|-----------|------|--------|
| E5.1 | WriteDB INSERT | create | Shipment record created | shipment_repo_test.go | Exists |
| E5.2 | WriteDB UPDATE | pause | Status updated to paused | shipment_repo_test.go | Exists |
| E5.3 | WriteDB UPDATE | resume | Status updated to active | shipment_repo_test.go | Exists |
| E5.4 | WriteDB UPDATE | complete | Status + completed_at | shipment_repo_test.go | Exists |
| E5.5 | CascadeGrove | assign_grove | Tasks receive grove_id | shipment_service_test.go | Exists |
| E5.6 | WriteDB DELETE | delete | Shipment deleted | shipment_repo_test.go | Exists |

---

## 6. Invariant Tests

| ID | Invariant | Property | Test Strategy | File | Status |
|----|-----------|----------|---------------|------|--------|
| I6.1 | id_format | IDs follow SHIP-XXX | Regex check | shipment_repo_test.go | Exists |
| I6.2 | id_unique | IDs are unique | Duplicate check | shipment_repo_test.go | Exists |
| I6.3 | status_valid | Status in (active, paused, complete) | Enum check | shipment_repo_test.go | Exists |
| I6.4 | mission_reference | Mission exists | FK constraint | shipment_repo_test.go | Exists |
| I6.5 | complete_not_pinned | Complete implies not pinned | Business logic | guards_test.go | New |

---

## 7. Test File Locations

| File | Purpose |
|------|---------|
| `internal/core/shipment/guards.go` | Guard implementations (NEW) |
| `internal/core/shipment/guards_test.go` | Guard unit tests (NEW) |
| `internal/app/shipment_service.go` | Service layer (EXISTS) |
| `internal/app/shipment_service_test.go` | Service tests (EXISTS) |
| `internal/adapters/sqlite/shipment_repo.go` | Repository (EXISTS) |
| `internal/adapters/sqlite/shipment_repo_test.go` | Repository tests (EXISTS) |

---

## 8. Implementation Checklist

### Context Structs to Create

- [x] `GuardResult` - shared from grove package pattern
- [ ] `CreateShipmentContext`
- [ ] `CompleteShipmentContext`
- [ ] `StatusTransitionContext`
- [ ] `AssignGroveContext`

### Guard Functions to Create

- [ ] `CanCreateShipment(ctx CreateShipmentContext) GuardResult`
- [ ] `CanCompleteShipment(ctx CompleteShipmentContext) GuardResult`
- [ ] `CanPauseShipment(ctx StatusTransitionContext) GuardResult`
- [ ] `CanResumeShipment(ctx StatusTransitionContext) GuardResult`
- [ ] `CanAssignGrove(ctx AssignGroveContext) GuardResult`

### Test Cases to Create (17 total)

**CanCreateShipment (2):**
- [ ] TestCanCreateShipment_MissionExists
- [ ] TestCanCreateShipment_MissionNotFound

**CanCompleteShipment (2):**
- [ ] TestCanCompleteShipment_Unpinned
- [ ] TestCanCompleteShipment_Pinned

**CanPauseShipment (3):**
- [ ] TestCanPauseShipment_Active
- [ ] TestCanPauseShipment_Paused
- [ ] TestCanPauseShipment_Complete

**CanResumeShipment (3):**
- [ ] TestCanResumeShipment_Paused
- [ ] TestCanResumeShipment_Active
- [ ] TestCanResumeShipment_Complete

**CanAssignGrove (4):**
- [ ] TestCanAssignGrove_Unassigned
- [ ] TestCanAssignGrove_SameShipment
- [ ] TestCanAssignGrove_OtherShipment
- [ ] TestCanAssignGrove_ShipmentNotFound

**GuardResult.Error (2):**
- [ ] TestGuardResult_Error_Allowed
- [ ] TestGuardResult_Error_Blocked

---

## 9. Comparison: Grove vs Shipment Guards

| Aspect | Grove | Shipment |
|--------|-------|----------|
| ORC-only guards | Yes (create, archive, restore) | No |
| Status guards | not_archived, is_archived | is_active, is_paused, not_pinned |
| Exclusivity guards | No | grove_not_assigned_elsewhere |
| Total guards | 8 | 5 |
| Total test cases | 24 | 17 |

---

*Derived from shipment-workflow.yaml FSM specification*
