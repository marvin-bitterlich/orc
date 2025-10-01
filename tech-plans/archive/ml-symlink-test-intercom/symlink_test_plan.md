# Symlink Architecture Test Plan

**Status**: investigating

## Problem & Solution
**Current Issue:** Testing the new repo root .tech-plans symlink architecture
**Solution:** Validate that tech plans created in ORC-managed location appear via symlink in worktree

## Testing Strategy

### Phase 1: Symlink Validation
- [x] Create .tech-plans symlink in repo root
- [x] Add .tech-plans to global gitignore
- [x] Create test tech plan in ORC location
- [ ] Verify plan appears in worktree via symlink
- [ ] Test /tech-plan command context awareness

### Phase 2: Workflow Integration  
- [ ] Test /bootstrap command integration
- [ ] Verify /janitor can manage these plans
- [ ] Test cross-worktree visibility from orchestrator

## Implementation Notes

This approach puts tech plans at repo root level (.tech-plans/) instead of buried in .claude/tech_plans, making them more discoverable while maintaining central ORC management.

The symlink points to: `/Users/looneym/src/orc/tech-plans/in-progress/ml-symlink-test-intercom/`