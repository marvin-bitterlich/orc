# Tech Planning Command

Generate focused technical planning documents using proven structure.

## Role

You're helping collaborators stay organized and deliberate while hacking on something together. Cut corporate ceremony - focus on clear thinking, functional code, and getting things done. Structure without bureaucracy.

## Usage

```
/tech-plan [PROJECT_NAME] [TYPE]
```

**Types:**
- `feature` - New feature development  
- `research` - Architecture/analysis work
- `debug` - Bug investigation

**Purpose**: Create NEW tech plans only. For managing existing plans (status updates, archiving), use `/janitor tech-plans`.

**File Storage**: Tech plans are context-aware:
- **Worktree Context**: Saved to `.tech-plans/project_name.md` (local to investigation)
- **ORC Context**: Saved to `tech-plans/backlog/project_name.md` for strategic planning

## Template Structure

```markdown
# [Project Name]

**Status**: investigating | in_progress | paused | done

## Problem & Solution
**Current Issue:** [What's broken/missing/inefficient]
**Solution:** [What we're building in one sentence]

## [Context Section]
[Current system state, requirements, or background]

## Implementation
### Approach
[High-level solution strategy]

### [Interface/API/Contract]
[Key interfaces, commands, or user interactions]

## Testing Strategy
[How we'll validate it works]

## Implementation Plan
### Phase 1: [Core/Foundation] 
### Phase 2: [Integration/Features]
### Phase 3: [Polish/Validation]
```

## Process

**NEW PLAN CREATION ONLY**:

1. **Discovery Conversation**: Understand what to build/fix/improve
2. **Create Plan File**: Context-aware location using template
3. **Initial Planning**: Fill out Problem & Solution, set status to "investigating"
4. **Live Documentation**: Update the plan file as we discuss approach
5. **Collaborative Refinement**: Explore alternatives, dive into technical details

**File Management**: 
- **Context Detection**: Check for `.tech-plans/` directory (worktree) vs `tech-plans/` dir (ORC)
- **Worktree**: Create in `.tech-plans/` (local to investigation, travels with the work)
- **ORC**: Create in `tech-plans/backlog/` for strategic planning  
- Start with status: "investigating" (updated lifecycle)
- For existing plan updates, use `/janitor tech-plans` instead

