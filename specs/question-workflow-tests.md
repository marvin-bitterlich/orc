# Question Workflow FSM - Test Matrix

**Generated:** 2026-01-20
**Source:** `specs/question-workflow.yaml`
**Context:** CON-003 FSM-First Testing Strategy (Slice 6)

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

### 1.1 CanCreateQuestion

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.1.1 | Mission exists, no investigation | MissionExists: true, InvestigationID: "" | Allowed: true | To implement |
| G1.1.2 | Mission exists with investigation | MissionExists: true, InvestigationID: "INV-001", InvestigationExists: true | Allowed: true | To implement |
| G1.1.3 | Mission not found | MissionExists: false | Allowed: false, "mission not found" | To implement |
| G1.1.4 | Investigation not found | MissionExists: true, InvestigationID: "INV-999", InvestigationExists: false | Allowed: false, "investigation not found" | To implement |

### 1.2 CanAnswerQuestion

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.2.1 | Open unpinned question | Status: "open", IsPinned: false | Allowed: true | To implement |
| G1.2.2 | Open pinned question | Status: "open", IsPinned: true | Allowed: false, "cannot answer pinned" | To implement |
| G1.2.3 | Already answered question | Status: "answered", IsPinned: false | Allowed: false, "can only answer open" | To implement |
| G1.2.4 | Answered and pinned (edge case) | Status: "answered", IsPinned: true | Allowed: false, "can only answer open" | To implement |

### 1.3 GuardResult.Error

| ID | Test Case | Input | Expected | Status |
|----|-----------|-------|----------|--------|
| G1.3.1 | Allowed result | Allowed: true | error == nil | To implement |
| G1.3.2 | Blocked result | Allowed: false, Reason: "test" | error.Error() == "test" | To implement |

---

## 2. Guard Tests Summary

| Guard Function | Test Count | Source Line |
|----------------|------------|-------------|
| CanCreateQuestion | 4 | question_service.go:27-45 |
| CanAnswerQuestion | 4 | NEW (missing guards in service) |
| GuardResult.Error | 2 | (helper method) |
| **Total** | **10** | |

---

## 3. Transition Tests (Service Level)

These tests verify end-to-end transitions through the service layer.

### 3.1 Happy Path Transitions

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T3.1.1 | create | initial | open | Create question with mission | question_service_test.go | Exists |
| T3.1.2 | answer | open | answered | Answer open question | question_service_test.go | Exists |

### 3.2 Self-Transitions

| ID | Transition | State | Test Case | File | Status |
|----|------------|-------|-----------|------|--------|
| T3.2.1 | update | open | Update title/description | question_service_test.go | Exists |
| T3.2.2 | pin | open | Pin question | question_service_test.go | Exists |
| T3.2.3 | unpin | open | Unpin question | question_service_test.go | Exists |

---

## 4. Guard Failure Tests (Negative Cases)

### 4.1 mission_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.1.1 | mission_exists | create | Create with missing mission | "mission not found" | guards_test.go |

### 4.2 investigation_exists_if_provided Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.2.1 | investigation_exists_if_provided | create | Create with missing investigation | "investigation not found" | guards_test.go |

### 4.3 is_open Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.3.1 | is_open | answer | Answer already answered question | "can only answer open" | guards_test.go |

### 4.4 not_pinned Guard

| ID | Guard | Transition | Test Case | Expected Error | File |
|----|-------|------------|-----------|----------------|------|
| F4.4.1 | not_pinned | answer | Answer pinned question | "cannot answer pinned" | guards_test.go |

---

## 5. Effect Tests

| ID | Effect | Transition | Test Case | File | Status |
|----|--------|------------|-----------|------|--------|
| E5.1 | WriteDB INSERT | create | Question record created | question_repo_test.go | Exists |
| E5.2 | WriteDB UPDATE | update | Title/description updated | question_repo_test.go | Exists |
| E5.3 | WriteDB UPDATE | answer | Status + answer + answered_at | question_repo_test.go | Exists |
| E5.4 | WriteDB DELETE | delete | Question deleted | question_repo_test.go | Exists |

---

## 6. Invariant Tests

| ID | Invariant | Property | Test Strategy | File | Status |
|----|-----------|----------|---------------|------|--------|
| I6.1 | id_format | IDs follow Q-XXX | Regex check | question_repo_test.go | Exists |
| I6.2 | id_unique | IDs are unique | Duplicate check | question_repo_test.go | Exists |
| I6.3 | status_valid | Status in (open, answered) | Enum check | question_repo_test.go | Exists |
| I6.4 | mission_reference | Mission exists | FK constraint | question_repo_test.go | Exists |
| I6.5 | investigation_reference | Investigation exists if set | FK constraint | question_repo_test.go | Exists |
| I6.6 | answered_not_pinned | Answered implies not pinned | Business logic | guards_test.go | New |

---

## 7. Test File Locations

| File | Purpose |
|------|---------|
| `internal/core/question/guards.go` | Guard implementations (NEW) |
| `internal/core/question/guards_test.go` | Guard unit tests (NEW) |
| `internal/app/question_service.go` | Service layer (EXISTS) |
| `internal/app/question_service_test.go` | Service tests (EXISTS) |
| `internal/adapters/sqlite/question_repo.go` | Repository (EXISTS) |
| `internal/adapters/sqlite/question_repo_test.go` | Repository tests (EXISTS) |

---

## 8. Implementation Checklist

### Context Structs to Create

- [x] `GuardResult` - shared pattern from other packages
- [ ] `CreateQuestionContext`
- [ ] `AnswerQuestionContext`

### Guard Functions to Create

- [ ] `CanCreateQuestion(ctx CreateQuestionContext) GuardResult`
- [ ] `CanAnswerQuestion(ctx AnswerQuestionContext) GuardResult`

### Test Cases to Create (10 total)

**CanCreateQuestion (4):**
- [ ] TestCanCreateQuestion_MissionExists_NoInvestigation
- [ ] TestCanCreateQuestion_MissionExists_WithInvestigation
- [ ] TestCanCreateQuestion_MissionNotFound
- [ ] TestCanCreateQuestion_InvestigationNotFound

**CanAnswerQuestion (4):**
- [ ] TestCanAnswerQuestion_OpenUnpinned
- [ ] TestCanAnswerQuestion_OpenPinned
- [ ] TestCanAnswerQuestion_AlreadyAnswered
- [ ] TestCanAnswerQuestion_AnsweredAndPinned

**GuardResult.Error (2):**
- [ ] TestGuardResult_Error_Allowed
- [ ] TestGuardResult_Error_Blocked

---

## 9. Comparison: Question vs Conclave/Investigation Guards

| Aspect | Conclave/Investigation | Question |
|--------|------------------------|----------|
| Parent entity | Mission (required) | Mission (required), Investigation (optional) |
| States | active, paused, complete | open, answered |
| Pause/Resume | Yes | No |
| Status guards | is_active, is_paused | is_open |
| Pinned guard | not_pinned (for complete) | not_pinned (for answer) |
| ID format | CON-XXX / INV-XXX | Q-XXX |
| Total guards | 4 | 2 |
| Total test cases | 12 | 10 |

---

*Derived from question-workflow.yaml FSM specification*
