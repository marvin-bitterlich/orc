# Grove Provisioning FSM - Test Matrix

**Generated:** 2026-01-20
**Source:** `specs/grove-provisioning.yaml`
**Context:** CON-003 FSM-First Testing Strategy (Slice 1)

---

## Test Categories

| Category | Description | Priority |
|----------|-------------|----------|
| Transition Tests | Happy path state transitions | P0 |
| Guard Failure Tests | Negative cases where guards block | P1 |
| Effect Tests | Side effects execute correctly | P2 |
| Invariant Tests | Property-based constraints | P3 |

---

## 1. Transition Tests (Happy Path)

### 1.1 Creation Transition

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T1.1.1 | create_grove | initial | active | ORC creates grove with existing mission | guards_test.go | ✅ Exists |
| T1.1.2 | create_grove | initial | active | ORC creates grove with repos | grove_service_test.go | ✅ Exists |

### 1.2 Open Transition (No State Change)

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T1.2.1 | open_grove | active | active | Open grove in TMux session | guards_test.go | ✅ Exists |

### 1.3 Rename Transition (No State Change)

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T1.3.1 | rename_grove | active | active | Rename existing grove | guards_test.go | ✅ Exists |

### 1.4 Archive Transition

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T1.4.1 | archive_grove | active | archived | ORC archives active grove | guards_test.go | ❌ **NEW** |

### 1.5 Restore Transition

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T1.5.1 | restore_grove | archived | active | ORC restores archived grove | guards_test.go | ❌ **NEW** |

### 1.6 Delete Transitions

| ID | Transition | From | To | Test Case | File | Status |
|----|------------|------|-----|-----------|------|--------|
| T1.6.1 | delete_active_grove | active | deleted | Delete grove with no active tasks | guards_test.go | ✅ Exists |
| T1.6.2 | delete_archived_grove | archived | deleted | Delete archived grove | guards_test.go | ❌ **NEW** (implied) |

---

## 2. Guard Failure Tests (Negative Cases)

### 2.1 is_orc Guard

| ID | Guard | Transition | Test Case | Expected Error | File | Status |
|----|-------|------------|-----------|----------------|------|--------|
| G2.1.1 | is_orc | create_grove | IMP creates grove | "IMPs cannot create groves..." | guards_test.go | ✅ Exists |
| G2.1.2 | is_orc | archive_grove | IMP archives grove | "IMPs cannot archive groves" | guards_test.go | ❌ **NEW** |
| G2.1.3 | is_orc | restore_grove | IMP restores grove | "IMPs cannot restore groves" | guards_test.go | ❌ **NEW** |

### 2.2 mission_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File | Status |
|----|-------|------------|-----------|----------------|------|--------|
| G2.2.1 | mission_exists | create_grove | Create for missing mission | "mission not found" | guards_test.go | ✅ Exists |

### 2.3 grove_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File | Status |
|----|-------|------------|-----------|----------------|------|--------|
| G2.3.1 | grove_exists | open_grove | Open non-existent grove | "grove not found" | guards_test.go | ✅ Exists |
| G2.3.2 | grove_exists | rename_grove | Rename non-existent grove | "grove not found" | guards_test.go | ✅ Exists |
| G2.3.3 | grove_exists | archive_grove | Archive non-existent grove | "grove not found" | guards_test.go | ❌ **NEW** |
| G2.3.4 | grove_exists | restore_grove | Restore non-existent grove | "grove not found" | guards_test.go | ❌ **NEW** |

### 2.4 path_exists Guard

| ID | Guard | Transition | Test Case | Expected Error | File | Status |
|----|-------|------------|-----------|----------------|------|--------|
| G2.4.1 | path_exists | open_grove | Open unmaterialized grove | "worktree not found" | guards_test.go | ✅ Exists |

### 2.5 in_tmux Guard

| ID | Guard | Transition | Test Case | Expected Error | File | Status |
|----|-------|------------|-----------|----------------|------|--------|
| G2.5.1 | in_tmux | open_grove | Open outside TMux | "not in a TMux session" | guards_test.go | ✅ Exists |

### 2.6 not_archived / is_archived Guards

| ID | Guard | Transition | Test Case | Expected Error | File | Status |
|----|-------|------------|-----------|----------------|------|--------|
| G2.6.1 | not_archived | archive_grove | Archive already-archived grove | "grove is already archived" | guards_test.go | ❌ **NEW** |
| G2.6.2 | is_archived | restore_grove | Restore active grove (not archived) | "grove is not archived" | guards_test.go | ❌ **NEW** |

### 2.7 no_tasks_or_force Guard

| ID | Guard | Transition | Test Case | Expected Error | File | Status |
|----|-------|------------|-----------|----------------|------|--------|
| G2.7.1 | no_tasks_or_force | delete_grove | Delete with active tasks | "has N active tasks" | guards_test.go | ✅ Exists |
| G2.7.2 | no_tasks_or_force | delete_grove | Delete with --force | (succeeds) | guards_test.go | ✅ Exists |

---

## 3. Effect Tests

| ID | Effect | Transition | Test Case | File | Status |
|----|--------|------------|-----------|------|--------|
| E3.1 | WriteDB INSERT | create_grove | Grove record created in DB | grove_repo_test.go | ✅ Exists |
| E3.2 | CreateWorkspace | create_grove | Grove directory created | planner_test.go | ✅ Exists |
| E3.3 | GitWorktreeAdd | create_grove | Git worktree added | planner_test.go | ✅ Exists |
| E3.4 | WriteConfig | create_grove | Config file written | planner_test.go | ✅ Exists |
| E3.5 | TmuxNewWindow | open_grove | TMux window created | (integration) | ⚪ Manual |
| E3.6 | WriteDB UPDATE | archive_grove | Status updated to archived | grove_repo_test.go | ❌ **NEW** |
| E3.7 | WriteDB UPDATE | restore_grove | Status updated to active | grove_repo_test.go | ❌ **NEW** |
| E3.8 | WriteDB DELETE | delete_grove | Grove record deleted | grove_repo_test.go | ✅ Exists |

---

## 4. Invariant Tests

| ID | Invariant | Property | Test Strategy | File | Status |
|----|-----------|----------|---------------|------|--------|
| I4.1 | id_format | IDs follow GROVE-XXX | Regex check on creation | grove_repo_test.go | ✅ Exists |
| I4.2 | id_unique | IDs are unique | Duplicate check | grove_repo_test.go | ✅ Exists |
| I4.3 | status_valid | Status in (active, archived) | Enum check | grove_repo_test.go | ✅ Exists |
| I4.4 | mission_reference | Mission exists | FK constraint | grove_repo_test.go | ✅ Exists |
| I4.5 | name_not_empty | Name length > 0 | Validation | grove_service_test.go | ✅ Exists |

---

## 5. Test Summary

### Existing Tests (16 guard tests)

| Guard Function | Test Count | Status |
|----------------|------------|--------|
| CanCreateGrove | 4 | ✅ Complete |
| CanOpenGrove | 5 | ✅ Complete |
| CanDeleteGrove | 4 | ✅ Complete |
| CanRenameGrove | 2 | ✅ Complete |
| GuardResult.Error | 1 | ✅ Complete |
| **Total** | **16** | |

### New Tests to Add (8 guard tests)

| Guard Function | Test Count | Status |
|----------------|------------|--------|
| CanArchiveGrove | 4 | ❌ To implement |
| CanRestoreGrove | 4 | ❌ To implement |
| **Total** | **8** | |

### Final Target

| Category | Count |
|----------|-------|
| Existing guard tests | 16 |
| New guard tests | 8 |
| **Total guard tests** | **24** |

---

## 6. Implementation Checklist

### Guards to Add

- [ ] `ArchiveGroveContext` struct
- [ ] `CanArchiveGrove(ctx ArchiveGroveContext) GuardResult`
- [ ] `CanRestoreGrove(ctx ArchiveGroveContext) GuardResult`

### Test Cases to Add

- [ ] TestCanArchiveGrove_ORCCanArchive
- [ ] TestCanArchiveGrove_IMPBlocked
- [ ] TestCanArchiveGrove_NotFound
- [ ] TestCanArchiveGrove_AlreadyArchived
- [ ] TestCanRestoreGrove_ORCCanRestore
- [ ] TestCanRestoreGrove_IMPBlocked
- [ ] TestCanRestoreGrove_NotFound
- [ ] TestCanRestoreGrove_NotArchived

---

## 7. Test File Locations

| File | Purpose |
|------|---------|
| `internal/core/grove/guards.go` | Guard implementations |
| `internal/core/grove/guards_test.go` | Guard unit tests |
| `internal/core/grove/planner.go` | Effect planning |
| `internal/core/grove/planner_test.go` | Planner tests |
| `internal/app/grove_service.go` | Service layer |
| `internal/app/grove_service_test.go` | Service tests |
| `internal/adapters/sqlite/grove_repo.go` | Repository |
| `internal/adapters/sqlite/grove_repo_test.go` | Repository tests |

---

*Derived from grove-provisioning.yaml FSM specification*
