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

The tech plan template is centralized in `templates/tech-plan.md` for consistency across all commands.

**Template Location**: `/Users/looneym/src/orc/templates/tech-plan.md`

**Template Loading**: Load the template content from the centralized location and customize with project-specific details.

## Process

**NEW PLAN CREATION ONLY**:

1. **Discovery Conversation**: Understand what to build/fix/improve
2. **Load Template**: Read centralized template from `templates/tech-plan.md`
3. **Create Plan File**: Context-aware location with customized template content
4. **Initial Planning**: Fill out Problem & Solution, set status to "investigating"
5. **Live Documentation**: Update the plan file as we discuss approach
6. **Collaborative Refinement**: Explore alternatives, dive into technical details

**File Management**: 
- **Template Loading**: `cat /Users/looneym/src/orc/templates/tech-plan.md` to get base template
- **Project Customization**: Replace `[PROJECT_NAME]` with actual project name
- **Context Detection**: Check for `.tech-plans/` directory (worktree) vs `tech-plans/` dir (ORC)
- **Worktree**: Create in `.tech-plans/` (local to investigation, travels with the work)
- **ORC**: Create in `tech-plans/backlog/` for strategic planning  
- Start with status: "investigating" (updated lifecycle)
- For existing plan updates, use `/janitor tech-plans` instead

**Template Usage**:
- **Read Template**: Use `@/Users/looneym/src/orc/templates/tech-plan.md` to include template content
- **Customize**: Replace `[PROJECT_NAME]` with actual project name in the file content
- **Create File**: Write customized content to appropriate context location

