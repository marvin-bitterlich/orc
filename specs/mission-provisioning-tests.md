# Mission Provisioning Test Matrix

Generated from: `specs/mission-provisioning.yaml`

## Test Categories

### 1. Transition Tests (Happy Path)

| Transition | From | To | Event | Guards | Test Cases |
|------------|------|-----|-------|--------|------------|
| create_mission | initial | active | create | is_orc | ORC creates mission with title only |
| | | | | | ORC creates mission with title and description |
| update_mission | active | active | update | - | Update title only |
| | | | | | Update description only |
| | | | | | Update both title and description |
| | | | | | Update with empty values (no-op) |
| start_mission | active | active | start | is_orc, not_in_orc_source | ORC starts mission, creates workspace |
| | | | | | ORC starts mission with custom workspace path |
| launch_mission | active | active | launch | is_orc, not_in_orc_source | Launch creates workspace and groves directory |
| | | | | | Launch moves groves to standard location |
| | | | | | Launch writes grove configs |
| | | | | | Launch with --tmux creates session |
| | | | | | Launch without --tmux skips TMux |
| | | | | | Launch is idempotent (run twice, same result) |
| pin_mission | active | active | pin | exists | Pin active mission |
| | | | | | Pin already-pinned mission (no-op) |
| unpin_mission | active | active | unpin | exists | Unpin pinned mission |
| | | | | | Unpin unpinned mission (no-op) |
| complete_mission | active | complete | complete | not_pinned | Complete unpinned mission |
| archive_active_mission | active | archived | archive | not_pinned | Archive active unpinned mission |
| archive_complete_mission | complete | archived | archive | not_pinned | Archive completed mission |
| pause_mission | active | paused | update | status_change_to_paused | Pause active mission |
| resume_mission | paused | active | update | status_change_to_active | Resume paused mission |
| delete_active_mission | active | deleted | delete | no_dependents_or_force | Delete mission with no dependents |
| | | | | | Delete mission with --force (succeeds, orphans dependents) |
| delete_archived_mission | archived | deleted | delete | no_dependents_or_force | Delete archived mission |

### 2. Guard Failure Tests (Negative Cases)

| Guard | Transition | Test Case | Expected Error |
|-------|------------|-----------|----------------|
| is_orc | create_mission | IMP blocked from creating mission | "IMPs cannot create missions - only ORC can create missions" |
| is_orc | start_mission | IMP blocked from starting mission | "IMPs cannot start missions - only ORC can start missions" |
| is_orc | launch_mission | IMP blocked from launching mission | "IMPs cannot launch missions - only ORC can launch missions" |
| not_in_orc_source | start_mission | Start from ORC source directory blocked | "Cannot run this command from ORC source directory" |
| not_in_orc_source | launch_mission | Launch from ORC source directory blocked | "Cannot run this command from ORC source directory" |
| not_pinned | complete_mission | Complete pinned mission (error) | "Cannot complete pinned mission {id}. Unpin first with: orc mission unpin {id}" |
| not_pinned | archive_active_mission | Archive pinned mission (error) | "Cannot archive pinned mission {id}. Unpin first with: orc mission unpin {id}" |
| no_dependents_or_force | delete_active_mission | Delete mission with groves (error without --force) | "Mission has {count} groves and {count} shipments. Use --force to delete anyway" |
| no_dependents_or_force | delete_active_mission | Delete mission with shipments (error without --force) | "Mission has {count} groves and {count} shipments. Use --force to delete anyway" |
| exists | pin_mission | Pin non-existent mission (error) | "Mission {id} not found" |

### 3. Edge Case Tests

| Category | Test Case | Expected Behavior |
|----------|-----------|-------------------|
| TMux | Start with existing TMux session | Error (session exists) |
| Idempotency | Launch same mission twice | Same result (idempotent) |
| No-op | Complete already-complete mission | No-op or error |
| No-op | Pin already-pinned mission | No-op |
| No-op | Unpin unpinned mission | No-op |

### 4. Invariant Tests (Property-Based)

| Invariant | Property | Test Strategy |
|-----------|----------|---------------|
| id_format | Mission IDs follow MISSION-XXX format | Regex check on all created missions |
| id_unique | Mission IDs are unique | Attempt duplicate creation |
| status_valid | Status is one of valid values | Attempt invalid status via direct DB |
| pinned_blocks_terminal | Pinned missions cannot be in terminal states | Attempt to pin then complete/archive |
| completed_has_timestamp | Complete status requires completed_at | Check completed_at after complete transition |
| timestamps_ordered | completed_at >= created_at | Check timestamp ordering after complete |

### 5. State Reachability Tests

| State | Reachable Via | Test |
|-------|---------------|------|
| initial | - | Default state before any mission exists |
| active | create from initial | Create mission, verify status = "active" |
| paused | pause from active | Pause mission, verify status = "paused" |
| complete | complete from active | Complete mission, verify status = "complete" |
| archived | archive from active/complete | Archive mission, verify status = "archived" |
| deleted | delete from active/archived | Delete mission, verify removed from DB |

---

## Test Implementation Priority

### P0 - Core Happy Path
1. create_mission (ORC creates with title)
2. complete_mission (unpinned)
3. archive_complete_mission
4. delete_archived_mission

### P1 - Guard Enforcement
1. is_orc guard (IMP blocked)
2. not_pinned guard (pinned blocks complete/archive)
3. no_dependents_or_force guard

### P2 - Infrastructure Provisioning
1. start_mission (workspace creation)
2. launch_mission (idempotent provisioning)
3. TMux session management

### P3 - Edge Cases & Invariants
1. Idempotency tests
2. No-op handling
3. Property-based invariant tests

---

## Mapping to Existing Tests

| FSM Test Case | Existing Shell Test | Status |
|---------------|---------------------|--------|
| ORC creates mission with title only | 01-test-mission-creation.sh:test_basic | ✅ Covered |
| ORC creates mission with title and description | 01-test-mission-creation.sh:test_with_description | ✅ Covered |
| Archive active unpinned mission | 01-test-mission-creation.sh:test_archive | ✅ Covered |
| IMP blocked from creating mission | - | ❌ Gap |
| Complete pinned mission (error) | - | ❌ Gap |
| Launch is idempotent | - | ❌ Gap |
| Pin/Unpin lifecycle | - | ❌ Gap |
| Delete with dependents guard | - | ❌ Gap |

## Test Gaps Summary

**High Priority Gaps:**
- Agent identity checks (is_orc guard) - no tests
- Pinned state transitions - no tests
- Deletion guards with dependents - no tests
- Launch idempotency - no tests

**Medium Priority Gaps:**
- Pause/resume workflow (rarely used but defined)
- TMux error cases
- Edge case no-ops
