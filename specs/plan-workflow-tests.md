# Plan Workflow FSM - Test Matrix

**Generated:** 2026-01-20
**Source:** `specs/plan-workflow.yaml`
**Context:** CON-003 FSM-First Testing Strategy (Slice 7)

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

### 1.1 CanCreatePlan

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.1.1 | Mission exists, no shipment | MissionExists: true, ShipmentID: "" | Allowed: true | To implement |
| G1.1.2 | Mission exists with shipment, no active plan | MissionExists: true, ShipmentID: "SHIP-001", ShipmentExists: true, ShipmentHasActivePlan: false | Allowed: true | To implement |
| G1.1.3 | Mission not found | MissionExists: false | Allowed: false, "mission not found" | To implement |
| G1.1.4 | Shipment not found | MissionExists: true, ShipmentID: "SHIP-999", ShipmentExists: false | Allowed: false, "shipment not found" | To implement |
| G1.1.5 | Shipment already has active plan | MissionExists: true, ShipmentID: "SHIP-001", ShipmentExists: true, ShipmentHasActivePlan: true | Allowed: false, "already has active plan" | To implement |

### 1.2 CanApprovePlan

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.2.1 | Draft unpinned plan | Status: "draft", IsPinned: false | Allowed: true | To implement |
| G1.2.2 | Draft pinned plan | Status: "draft", IsPinned: true | Allowed: false, "cannot approve pinned" | To implement |
| G1.2.3 | Already approved plan | Status: "approved", IsPinned: false | Allowed: false, "can only approve draft" | To implement |
| G1.2.4 | Approved and pinned (edge case) | Status: "approved", IsPinned: true | Allowed: false, "can only approve draft" | To implement |

### 1.3 CanDeletePlan

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.3.1 | Not pinned | IsPinned: false | Allowed: true | To implement |
| G1.3.2 | Pinned | IsPinned: true | Allowed: false, "cannot delete pinned" | To implement |
| G1.3.3 | Approved not pinned | IsPinned: false | Allowed: true | To implement |

### 1.4 GuardResult.Error

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.4.1 | Allowed result | Allowed: true | error == nil | To implement |
| G1.4.2 | Blocked result | Allowed: false, Reason: "test" | error.Error() == "test" | To implement |

---

## 2. Guard Tests Summary

| Guard Function | Test Count | Source Line |
|----------------|------------|-------------|
| CanCreatePlan | 5 | internal/core/plan/guards.go |
| CanApprovePlan | 4 | internal/core/plan/guards.go |
| CanDeletePlan | 3 | internal/core/plan/guards.go |
| GuardResult.Error | 2 | (helper method) |
| **Total** | **14** | |

---

## 3. Transition Tests (Service Level)

These tests verify end-to-end transitions through the service layer.

### 3.1 Happy Path Transitions

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T3.1.1 | create | initial | draft | Create plan with mission | plan_service_test.go | Exists |
| T3.1.2 | approve | draft | approved | Approve draft plan | plan_service_test.go | Exists |

### 3.2 Self-Transitions

| ID | Transition | State | Test Case | File | Status |
|----|------------|-------|-----------|------|--------|
| T3.2.1 | update | draft | Update title/description | plan_service_test.go | Exists |
| T3.2.2 | pin | draft | Pin plan | plan_service_test.go | Exists |
| T3.2.3 | unpin | draft | Unpin plan | plan_service_test.go | Exists |

---

## 4. Guard Failure Tests (Negative Cases)

### 4.1 mission_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.1.1 | mission_exists | create | Create with missing mission | "mission not found" | guards_test.go |

### 4.2 shipment_exists_if_provided Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.2.1 | shipment_exists_if_provided | create | Create with missing shipment | "shipment not found" | guards_test.go |

### 4.3 no_active_plan Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.3.1 | no_active_plan | create | Create when shipment has active plan | "already has active plan" | guards_test.go |

### 4.4 is_draft Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.4.1 | is_draft | approve | Approve already approved plan | "can only approve draft" | guards_test.go |

### 4.5 not_pinned Guard (approve)

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.5.1 | not_pinned | approve | Approve pinned plan | "cannot approve pinned" | guards_test.go |

### 4.6 not_pinned Guard (delete)

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.6.1 | not_pinned | delete | Delete pinned plan | "cannot delete pinned" | guards_test.go |

---

## 5. Effect Tests

| ID | Effect | Transition | Test Case | File | Status |
|----|--------|------------|-----------|------|--------|
| E5.1 | WriteDB INSERT | create | Plan record created | plan_repo_test.go | Exists |
| E5.2 | WriteDB UPDATE | update | Title/description updated | plan_repo_test.go | Exists |
| E5.3 | WriteDB UPDATE | approve | Status + approved_at | plan_repo_test.go | Exists |
| E5.4 | WriteDB DELETE | delete | Plan deleted | plan_repo_test.go | Exists |

---

## 6. Invariant Tests

| ID | Invariant | Property | Test Strategy | File | Status |
|----|-----------|----------|---------------|------|--------|
| I6.1 | id_format | IDs follow PLAN-XXX | Regex check | plan_repo_test.go | Exists |
| I6.2 | id_unique | IDs are unique | Duplicate check | plan_repo_test.go | Exists |
| I6.3 | status_valid | Status in (draft, approved) | Enum check | plan_repo_test.go | Exists |
| I6.4 | mission_reference | Mission exists | FK constraint | plan_repo_test.go | Exists |
| I6.5 | shipment_reference | Shipment exists if set | FK constraint | plan_repo_test.go | Exists |
| I6.6 | approved_not_pinned | Approved implies not pinned | Business logic | guards_test.go | New |
| I6.7 | one_draft_per_shipment | Max 1 draft per shipment | Business logic | guards_test.go | New |

---

## 7. Test File Locations

| File | Purpose |
|------|---------|
| `internal/core/plan/guards.go` | Guard implementations (NEW) |
| `internal/core/plan/guards_test.go` | Guard unit tests (NEW) |
| `internal/app/plan_service.go` | Service layer (EXISTS) |
| `internal/app/plan_service_test.go` | Service tests (EXISTS) |
| `internal/adapters/sqlite/plan_repo.go` | Repository (EXISTS) |
| `internal/adapters/sqlite/plan_repo_test.go` | Repository tests (EXISTS) |

---

## 8. Implementation Checklist

### Context Structs to Create

- [x] `GuardResult` - shared pattern from other packages
- [ ] `CreatePlanContext`
- [ ] `ApprovePlanContext`
- [ ] `DeletePlanContext`

### Guard Functions to Create

- [ ] `CanCreatePlan(ctx CreatePlanContext) GuardResult`
- [ ] `CanApprovePlan(ctx ApprovePlanContext) GuardResult`
- [ ] `CanDeletePlan(ctx DeletePlanContext) GuardResult`

### Test Cases to Create (14 total)

**CanCreatePlan (5):**
- [ ] TestCanCreatePlan_MissionExists_NoShipment
- [ ] TestCanCreatePlan_MissionExists_WithShipment_NoActivePlan
- [ ] TestCanCreatePlan_MissionNotFound
- [ ] TestCanCreatePlan_ShipmentNotFound
- [ ] TestCanCreatePlan_ShipmentHasActivePlan

**CanApprovePlan (4):**
- [ ] TestCanApprovePlan_DraftUnpinned
- [ ] TestCanApprovePlan_DraftPinned
- [ ] TestCanApprovePlan_AlreadyApproved
- [ ] TestCanApprovePlan_ApprovedAndPinned

**CanDeletePlan (3):**
- [ ] TestCanDeletePlan_NotPinned
- [ ] TestCanDeletePlan_Pinned
- [ ] TestCanDeletePlan_ApprovedNotPinned

**GuardResult.Error (2):**
- [ ] TestGuardResult_Error_Allowed
- [ ] TestGuardResult_Error_Blocked

---

## 9. Comparison: Plan vs Question Guards

| Aspect | Question | Plan |
|--------|----------|------|
| Parent entity | Mission (required), Investigation (optional) | Mission (required), Shipment (optional) |
| States | open, answered | draft, approved |
| Pause/Resume | No | No |
| Status guards | is_open | is_draft |
| Pinned guard | not_pinned (for answer) | not_pinned (for approve AND delete) |
| Delete guard | None | not_pinned |
| Unique constraint | None | Max 1 draft per shipment |
| ID format | Q-XXX | PLAN-XXX |
| Total guards | 2 | 3 |
| Total test cases | 10 | 14 |

---

*Derived from plan-workflow.yaml FSM specification*
