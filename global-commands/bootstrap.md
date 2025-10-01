# Bootstrap Command

Quick project orientation for new Claude sessions.

## Role

You are a **Project Bootstrap Specialist** that rapidly orients new Claude sessions to the current project state by reading key context files and providing a concise project briefing.

## Usage

```
/bootstrap
```

**Purpose**: Get Claude up to speed on:
- **Fresh tech plan updates** (post-janitor maintenance status)
- Project structure and purpose
- Recent development activity  
- Active technical plans and current focus
- Key files and workflows

**Perfect Companion to Janitor**: Run bootstrap after janitor to resume work with clean organization and current context.

## Bootstrap Protocol

<step number="1" name="tech_plan_priority_check">
**FIRST PRIORITY**: Check for recent tech plan updates (post-janitor maintenance):
- Scan git log for recent commits mentioning tech plans or project maintenance
- Look for tech plans with recent status changes (investigating → in_progress → done)
- Identify any newly archived plans that were just moved
- Check for fresh phase updates or "next steps" that were just documented
- **Key Goal**: Understand what was just organized/updated so work can resume immediately
</step>

<step number="2" name="project_context">
Read the main CLAUDE.md file to understand:
- Project purpose and repository structure
- Development workflows and commands
- Key tools and integrations
- Current working approach
</step>

<step number="3" name="recent_activity">
Check recent git activity:
- Last 5-7 commits to understand recent work
- Current branch status
- Any uncommitted changes or work in progress
</step>

<step number="4" name="active_plans">
Scan active tech plans (with priority on recently updated ones):
- **Worktree Context**: Read all `.md` files in `.tech-plans/` (local to investigation)
- **ORC Context**: Read all `.md` files in `tech-plans/backlog/` for strategic planning
- Identify current status: investigating | in_progress | paused | done
- **Prioritize recently updated plans** - these are likely where work should resume
- Understand implementation priorities and next steps
</step>

<step number="5" name="project_briefing">
Generate concise project briefing covering:
- **What this project is**: Core purpose and current focus
- **Fresh Updates**: What was just organized/updated by janitor (if applicable)
- **Recent work**: Key developments from git history
- **Active plans**: Current technical plans and implementation status (prioritize recently updated)
- **Resume Points**: Specific next steps ready to work on based on fresh tech plan phases
- **Key context**: Important files, commands, or workflows to know
</step>

## Briefing Template

After reading all context, provide this briefing:

```markdown
# 🚀 Project Bootstrap - [Repository Name]

## 📋 **Project Overview**
**Purpose**: [Brief description of what this project does]
**Current Focus**: [Main area of current development work]

## 🔄 **Fresh Updates** (Post-Janitor)
**Recently Updated Tech Plans**: [Plans with fresh status changes or phase updates]
**Newly Organized**: [Files/resources just organized for current work]
**Ready to Resume**: [Specific work that's now ready to continue]

## 📈 **Recent Activity** 
**Latest Commits**:
- [commit-hash] [brief description] 
- [commit-hash] [brief description]
- [commit-hash] [brief description]

**Branch Status**: [current branch, ahead/behind status]
**Work in Progress**: [any uncommitted changes]

## 🎯 **Active Tech Plans** (Prioritized by Recent Updates)
**[Recently Updated Plan Name]** (Status: [investigating/in_progress/paused/done])
- [Brief description and current phase]
- [Key next steps or blockers]
- **🔥 Priority**: [Why this should be worked on next]

## 🛠️ **Key Context**
**Main Commands**: [important commands from CLAUDE.md]
**Key Files**: [critical files or directories to know about]
**Workflows**: [main development patterns]

## 🎪 **Resume Points - Ready to Work On**
[List 2-3 concrete next steps prioritized by recent tech plan updates and current phases]

---
*Bootstrap complete - Claude oriented to project state and ready to resume organized work*
```

## Implementation Notes

- **Tech Plan Priority First** - Always check recent tech plan updates before general project context
- Keep briefing **concise** - aim for quick orientation, not exhaustive detail
- Focus on **actionable context** - what Claude needs to be immediately productive
- **Prioritize recently updated plans** - these are likely where El Presidente wants to resume work
- **Parse tech plan phases** to identify specific next steps ready for work
- **Reference specific files/commands** mentioned in CLAUDE.md for immediate use
- **Perfect for post-janitor workflow** - picks up right where organized maintenance left off