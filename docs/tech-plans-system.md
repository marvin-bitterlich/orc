# Tech Plans System

**Lightweight 3-State Planning for Quick Development Planning**

## Overview

The tech plans system provides a lightweight approach to planning development work without ceremony or complex project management overhead. Plans flow through three simple states and are organized for both individual focus and cross-project coordination.

## Design Philosophy

### Lightweight Planning
- **Goal**: "Quickly cutting plans for things I'm working on"
- **Anti-Pattern**: Complex lifecycle tracking and project management ceremony  
- **Focus**: Individual developer workflow, not team coordination complexity

### Simple State Model
- **investigating**: Figuring out the problem/approach
- **in_progress**: Actively working  
- **paused**: Blocked or deprioritized
- **done**: Completed (moves to archive)

### Simplified Ownership
- **Local Ownership**: Each worktree contains its own tech plans locally
- **Strategic Planning**: ORC backlog for future work and coordination
- **Historical Context**: Archive preserves completed work for reference

## Directory Structure

```
# Simplified Architecture - Plans Live With Work, Managed Locally

# Individual worktree tech plans (complete local ownership)
~/src/worktrees/ml-feature-intercom/.tech-plans/
├── investigation.md         # Active plans in root for easy access
├── implementation.md        # Current work items
├── backlog/                 # Local future/paused work  
│   ├── performance_optimization.md     # Future work for this investigation
│   └── error_handling_improvements.md  # Paused local work
└── archive/                 # Local completed work
    ├── initial_research.md             # Completed phases
    └── proof_of_concept.md             # Archived work

# ORC central planning (strategic coordination only)
orc/tech-plans/
├── backlog/                 # System-wide strategic planning
│   ├── orc_ecosystem_refinement.md     # Cross-system work
│   └── WO-004_perfbot_system_enhancements.md  # Strategic initiatives
└── archive/                 # Completed strategic work
    └── WO-012_dlq_bot_foundations.md   # Finished strategic work
```

## Tech Plan Template

### Basic Structure
```markdown
# [Plan Name]

**Status**: investigating | in_progress | paused | done

## Problem & Solution
**Current Issue:** [What's broken/missing/inefficient]
**Solution:** [What we're building in one sentence]

## Implementation
### Approach
[High-level solution strategy]

### Key Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

## Testing Strategy
[How we'll validate it works]

## Notes
[Implementation notes, discoveries, links]
```

### Status Lifecycle
1. **investigating**: Created with basic problem/solution, gathering information
2. **in_progress**: Implementation started, tasks being worked
3. **paused**: Work stopped (blocked, deprioritized, or waiting)  
4. **done**: Work completed → moves to archive

## Integration with Worktrees

### Local Access Pattern
```bash
# From any worktree - complete local management
ls .tech-plans/                    # Active plans for immediate work
ls .tech-plans/backlog/            # Future/paused work for this investigation  
ls .tech-plans/archive/            # Completed work for reference
vim .tech-plans/feature-analysis.md  # Edit active plan
```

### Complete Local Architecture  
```bash
# Full local ownership - no external dependencies
worktree/.tech-plans/              # Real directory, not symlink
├── feature-analysis.md           # Active work (root level for easy access)
├── implementation-notes.md       # Current focus items
├── backlog/                      # Local future work
│   └── optimization-ideas.md     # Ideas for later
└── archive/                      # Local completed work  
    └── initial-research.md       # Finished phases
```

### Creation Workflow
```bash
# From within worktree
/tech-plan feature-analysis        # Creates .tech-plans/feature-analysis.md
                                  # Actually stored locally, travels with work
```

## Command Integration

### `/tech-plan` Command
- **Context Detection**: Recognizes worktree vs ORC context
- **Automatic Placement**: Creates plans in correct location
- **Template Application**: Uses lightweight 4-state template

### `/bootstrap` Command  
- **Local Plans**: Reads from `.tech-plans/` directory and subdirectories
- **Context Loading**: Combines local tech plans with git history
- **Priority Focus**: Shows active plans (root level) and recent activity

### `/janitor` Command (Local Focus)
- **Local Lifecycle Management**: Manages tech plan states based on work evidence
- **Local Organization**: Organizes loose files and maintains `.tech-plans/` structure
- **Forest Clearing Guardian**: Keeps individual worktrees focused and organized

## Lifecycle Management

### State Transitions

#### investigating → in_progress
```bash
# Edit plan to update status and add implementation details
vim .tech-plans/plan-name.md
# Change: **Status**: investigating → in_progress
```

#### in_progress → paused
```bash
# Update status and note reason for pausing
vim .tech-plans/plan-name.md
# Change: **Status**: in_progress → paused
# Add notes about why paused and what's needed to resume
```

#### in_progress → done  
```bash
# Complete the work, then archive locally
vim .tech-plans/plan-name.md
# Change: **Status**: in_progress → done

# Archive locally via janitor or manual move
mv .tech-plans/plan-name.md .tech-plans/archive/
```

#### paused → backlog
```bash  
# Move paused work to local backlog to reduce active noise
mv .tech-plans/paused-plan.md .tech-plans/backlog/
```

### Local Focus Management

#### Active → Backlog (Reduce Noise)
```bash
# Move future/paused work to backlog to maintain active focus
mv .tech-plans/future-work.md .tech-plans/backlog/
# Keep only actively worked plans in .tech-plans/ root
```

#### Backlog → Active (Resume Work)
```bash
# Bring backlogged work back to active when ready
mv .tech-plans/backlog/ready-to-work.md .tech-plans/
# Now visible in active plans for immediate work
```

## Migration from Work Orders

### Conversion Pattern
Work orders have been migrated to tech plans structure:

- **01-backlog + 02-next + 03-in-progress** → `backlog/`
- **04-complete** → `archive/`
- **Complex lifecycle fields** → Simplified 4-state approach

### Template Simplification
```markdown
# Before (Work Order)
**Created**: 2025-08-21  
**Category**: 🤖 Automation  
**Priority**: Medium  
**Effort**: L  
**IMP Assignment**: Unassigned

# After (Tech Plan)
**Status**: investigating
```

## Best Practices

### Plan Creation
- **Start Simple**: Just problem/solution initially
- **Iterate**: Add details as understanding develops
- **Stay Focused**: One investigation per plan
- **Link Context**: Reference issues, PRs, documentation

### State Management
- **Honest Status**: Update status when reality changes
- **Clear Pausing**: Note why work stopped and what's needed to resume
- **Archive Promptly**: Move completed work to keep in-progress clean

### Cross-Worktree Coordination
- **Orchestrator View**: Use ORC perspective for status across all work
- **Individual Focus**: Use worktree symlinks for local work
- **Dependencies**: Note cross-plan dependencies when they exist

## Troubleshooting

### Plans Not Appearing in Worktree
```bash
# Check symlink
ls -la .tech-plans
# Should point to: orc/tech-plans/in-progress/[worktree-name]

# Check ORC directory exists
ls orc/tech-plans/in-progress/[worktree-name]/
```

### `/tech-plan` Command Issues
- **Wrong Location**: Verify command detects worktree context correctly
- **Permission Issues**: Ensure ORC directory is writable
- **Template Problems**: Check command uses simplified 4-state template

### Cross-Worktree Visibility
```bash
# Orchestrator view of all active work
ls orc/tech-plans/in-progress/*/
# Should show all worktree namespaces and their plans
```