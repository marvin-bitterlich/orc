# Tech Plan Command Modernization

**Status**: investigating

## Problem & Solution
**Current Issue:** The `/tech-plan` command still references old `.claude/tech_plans/` paths and outdated status values, preventing it from working with our new symlinked tech plans architecture.
**Solution:** Update the command to be context-aware and use the new 4-state lifecycle system.

## Context
The `/tech-plan` command is a core part of our universal command system but hasn't been updated to work with:
- **New architecture**: `.tech-plans/` symlinks in worktrees vs `tech-plans/backlog/` in ORC
- **Updated lifecycle**: `investigating|in_progress|paused|done` instead of `research|implementation|done`
- **Context awareness**: Should detect worktree vs ORC context and create plans accordingly

## Implementation
### Approach
Update the command to:
1. **Detect context**: Check for `.tech-plans/` symlink (worktree) vs `tech-plans/` directory (ORC)
2. **Context-aware creation**: 
   - Worktree: Create in `.tech-plans/` (stored via symlink in ORC)
   - ORC: Create in `tech-plans/backlog/` for strategic planning
3. **Modern lifecycle**: Use 4-state system with "investigating" as initial status
4. **Updated template**: Reflect new status values and simplified structure

### Command Logic Flow
```bash
# Context detection
if [ -L ".tech-plans" ]; then
    # Worktree context - create in symlinked directory
    LOCATION=".tech-plans/"
    CONTEXT="worktree"
elif [ -d "tech-plans" ]; then
    # ORC context - create in backlog
    LOCATION="tech-plans/backlog/"  
    CONTEXT="orc"
else
    # Unknown context - error or default behavior
    ERROR="Cannot determine context"
fi
```

### Template Updates
- Status values: `investigating|in_progress|paused|done`
- Initial status: `investigating`
- Context-appropriate file locations
- Simplified structure matching our lightweight approach

## Testing Strategy
1. **Worktree Testing**: Create tech plan from within a worktree, verify it appears in symlinked directory
2. **ORC Testing**: Create tech plan from ORC context, verify it goes to backlog
3. **File Verification**: Confirm plans are accessible from both local and central perspectives
4. **Integration Testing**: Verify compatibility with `/bootstrap` and `/janitor` commands

## Implementation Plan

### Phase 1: Command Logic Update
- [ ] Update context detection logic
- [ ] Implement context-aware file creation
- [ ] Update status lifecycle values
- [ ] Test basic functionality in both contexts

### Phase 2: Template Modernization  
- [ ] Update template with new status values
- [ ] Simplify structure for lightweight approach
- [ ] Ensure compatibility with existing tech plans
- [ ] Update documentation strings

### Phase 3: Integration Validation
- [ ] Test with `/bootstrap` command context loading
- [ ] Test with `/janitor` lifecycle management
- [ ] Verify symlink behavior works correctly
- [ ] Update any dependent documentation

## Notes
This is a critical update since `/tech-plan` is one of our core universal commands. The fix will enable proper tech plan creation in our new architecture and ensure consistency with the 4-state lifecycle we've established.

The command should maintain its "lightweight without ceremony" philosophy while gaining context awareness for our sophisticated worktree system.