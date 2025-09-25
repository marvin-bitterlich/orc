# CLAUDE.md - ORC Ecosystem Command Center

This repository serves as the central command and coordination layer for the ORC development ecosystem, managing universal commands, agents, and worktree orchestration.

## ORC Ecosystem Structure

```
/Users/looneym/src/orc/                    # ORC Command Center
├── CLAUDE.md                              # This file - ecosystem documentation
├── .claude/
│   ├── commands/                          # Master command definitions
│   │   ├── analyze-prompt.md             # Universal prompt analysis
│   │   ├── bootstrap.md                  # Universal project bootstrap
│   │   ├── janitor.md                    # Universal maintenance
│   │   └── tech-plan.md                  # Universal technical planning
│   ├── agents/                           # Universal agent definitions (planned)
│   └── tech_plans/                       # Strategic planning documents
├── spellbook/                            # Detailed procedures and knowledge base
└── work-trees -> /Users/looneym/src/worktrees   # Symlink to active worktrees

~/.claude/commands/                        # Global command access (symlinks to ORC masters)
├── analyze-prompt.md -> /Users/looneym/src/orc/.claude/commands/analyze-prompt.md
├── bootstrap.md -> /Users/looneym/src/orc/.claude/commands/bootstrap.md
├── janitor.md -> /Users/looneym/src/orc/.claude/commands/janitor.md
└── tech-plan.md -> /Users/looneym/src/orc/.claude/commands/tech-plan.md
```

## Universal Command System

### ORC-Managed Commands
All universal project management commands are centrally managed in ORC and made globally available through symlinks:

- **`/analyze-prompt`** - Advanced prompt quality assessment using latest Anthropic practices
- **`/bootstrap`** - Universal project initialization and setup
- **`/janitor`** - Complete project maintenance (CLAUDE.md validation, tech plan lifecycle, cleanup)
- **`/tech-plan`** - Structured technical planning with proven templates

### Command Management
- **Master Definitions**: Stored in `/Users/looneym/src/orc/.claude/commands/`
- **Global Access**: Symlinked from `~/.claude/commands/` for universal availability
- **Version Control**: All commands tracked in ORC repository for consistency
- **Updates**: Edit master files in ORC, changes propagate globally through symlinks

## Git Worktree Workflow

### Core Concept

Instead of switching between repositories and managing context across different directories, we create isolated feature environments where all related repositories are checked out as worktrees under a single feature directory.

### Automated Development Environment

Each worktree includes:
- **CLAUDE.md**: Context file documenting the investigation/feature (with worktree template)
- **TMux Integration**: Automated window setup with proper layout  
- **Development Tools**: Vim + NERDTree + Claude setup
- **Clean Isolation**: Worktree-claudes only know about their specific investigation

### Setting Up a New Investigation/Feature

#### Orchestrator Claude Workflow
**CRITICAL**: Orchestrator Claude must always create worktrees from latest origin/master WITHOUT disturbing current work:

```bash
# 1. Fetch latest origin/master WITHOUT disturbing current branch
cd /Users/looneym/src/infrastructure && git fetch origin
cd /Users/looneym/src/intercom && git fetch origin

# 2. Create worktrees directly from origin/master (preserves current work in main repos)
git worktree add /Users/looneym/src/worktrees/ml-feature-name/infrastructure -b ml/infrastructure-feature-name origin/master
git worktree add /Users/looneym/src/worktrees/ml-feature-name/intercom -b ml/intercom-feature-name origin/master

# 3. Add worktree template to CLAUDE.md
# 4. Set up tmux development environment
tmux new-window -n "feature-name" -c "/Users/looneym/src/worktrees/ml-feature-name" \; send-keys "muxup" Enter
```

#### Manual Method (Legacy)
```bash
# Create feature directory
mkdir -p worktrees/ml-feature-name

# Fetch latest origin/master without disturbing current work
cd ../../intercom && git fetch origin
cd ../infrastructure && git fetch origin  
cd ../muster && git fetch origin

# Add worktrees directly from origin/master
cd worktrees/ml-feature-name
git worktree add intercom ../../intercom -b ml/intercom-feature-name origin/master
git worktree add infrastructure ../../infrastructure -b ml/infrastructure-feature-name origin/master
git worktree add muster ../../muster -b ml/muster-feature-name origin/master
```

#### 2. Automated TMux Workflow
For investigations requiring only the Intercom app:

```bash
# Claude creates fully automated setup:
tmux new-window -n "feature-name" -c "/Users/looneym/src/worktrees/ml-feature-name" \; send-keys "muxup" Enter
```

This automatically:
- Creates named TMux window with correct working directory
- Sets up 3-pane layout: Vim (left) | Claude (top-right) | Shell (bottom-right)  
- Opens vim with CLAUDE.md file and NERDTree sidebar
- Starts Claude session in top-right pane

### Checking Active Work

#### List All Active Worktrees
```bash
# From any repo directory, see all active worktrees
cd intercom && git worktree list
```

#### Check Current Investigations
```bash
# See what feature investigations are active
ls -la worktrees/
```

#### Orchestrator Claude Integration
When working with Orchestrator Claude on worktrees:

**Check Active Work:**
- *"What are my active worktrees?"* → Claude lists directory names only from `worktrees/` folder
- *"What worktrees do we have active?"* → Same as above - just the directory names, no deep inspection

**Start Working:**
- *"I want to work on [feature-name]"* → Orchestrator Claude opens TMux window with `muxup` command:
  ```bash
  tmux new-window -n "feature-name" -c "/Users/looneym/src/worktrees/[feature-name]" \; send-keys "muxup" Enter
  ```
- *"Create new investigation for [topic]"* → Orchestrator Claude sets up new worktree with CLAUDE.md template

**Development Support:**
- Orchestrator Claude coordinates by updating worktree `CLAUDE.md` files (appearing as El Presidente communications)
- Each worktree-claude works independently on their specific investigation
- Worktree-claudes don't know about Orchestrator Claude - clean separation
- Context files include links to relevant Slack channels, PRs, and monitoring dashboards

**GitHub Issue Worktree Creation Workflow:**

When El Presidente provides a GitHub issue URL and requests a worktree setup:

1. **Fetch Issue Details**: Use `gh issue view <number> --repo intercom/intercom` to get full issue context
2. **Read All Comments**: Review issue description AND all comments for complete context
3. **Create Descriptive Name**: Generate meaningful branch/directory name based on actual problem (not just issue number)
4. **Set Up Worktrees**: Create from latest master with proper branch naming
5. **Comprehensive CLAUDE.md**: Include:
   - Problem summary from issue AND comments
   - All relevant links and resources mentioned
   - Investigation plan based on full context
   - Progress tracking structure

**Worktree CLAUDE.md Template:**
```markdown
# Investigation: [Descriptive Name Based on Issue]

## Environment Setup
You are working on a feature that spans multiple repositories. You have access to:

- **intercom/**: Main application repository (branch: `ml/descriptive-name`)
- **infrastructure/**: Terraform infrastructure configuration (branch: `ml/descriptive-name`)

Both repos are checked out from the latest master and ready for development.

## Your Mission
[Full problem description from issue + comments]

**GitHub Issue**: [URL]

### Problem Summary
[Detailed context from issue description and all comments]

## Available Resources
- **GitHub Issue**: [URL with full context]
- [All links, resources, monitoring tools mentioned in issue/comments]

## Status Update Protocol
**CRITICAL**: Update the "Current Status" section immediately when taking these actions:

### Mandatory Status Update Triggers:
1. **🔄 → 🟡 Investigation Complete**: After mapping the problem and creating GitHub issue
2. **🟡 → 🟠 Implementation Started**: When beginning code changes
3. **🟠 → 🟢 PRs Created**: When opening pull requests (include PR links)
4. **🟢 → ✅ Complete**: When work is finished/merged
5. **❌ Blocked**: When encountering blockers that halt progress

### Status Update Method:
Use this exact format in the "Current Status" section:
```
### Current Status
[EMOJI] **[STATUS_NAME]** - [Brief description]

**Key Actions Completed**:
- [Timestamp] [Action description]
- [Timestamp] [Action description]

**Active PRs**: [Links if any]
**Next Steps**: [What's needed to progress]
```

### Optional: Sub-Agent Status Updates
For complex status updates, use the Task tool with this specialized agent:
- **Agent Type**: `general-purpose`
- **Prompt**: "You are a status update specialist. Review the current investigation context and update the CLAUDE.md Current Status section with accurate progress. Focus on: completed actions, active PRs, blockers, and clear next steps."

## Progress
**IMPORTANT**: Keep this section updated as you work. Record your findings, decisions, and next steps.

### Current Status
[Current state based on issue comments]

### Investigation Plan  
[Tasks derived from issue discussion]

### Progress Log
[Timestamped entries of work completed and findings]
```

**Key Commands:**
- **List worktrees**: Just show directory names from `ls worktrees/` - no file reading
- **Open worktree**: Use TMux command to launch `muxup` in the worktree directory  
- **Status report**: Check worktree progress, git status, and recent commits

### Orchestrator Claude Boundaries
**CRITICAL SAFETY CHECK**: If El Presidente asks Orchestrator Claude to work directly on investigation code or make changes within a worktree:

```
❌ "Hey, can you fix that bug in the ingestion worker code?"
❌ "Can you update the CPU multiplier config?"  
❌ "Debug that error in the infrastructure/"

✅ Response: "El Presidente, I'm Orchestrator Claude - I coordinate worktrees but don't work directly on investigations. You'll want to switch to the worktree tmux window to work with the worktree-claude on that task."
```

**Orchestrator Claude ONLY does:**
- Worktree setup and management
- Status reporting across worktrees  
- Coordination and CLAUDE.md updates
- TMux window management

**Orchestrator Claude NEVER does:**
- Code changes within worktrees
- Investigation work
- Debugging specific technical issues
- Direct file edits in repos (except CLAUDE.md coordination)

### Status Reporting
When El Presidente asks for a status report on a worktree, Orchestrator Claude should:

1. **Check CLAUDE.md Progress**: Read the Progress section to see what worktree-claude has reported
2. **Git Status Check**: Run `git status` in each repo to see uncommitted changes
3. **Recent Commits**: Check `git log --oneline -5` to see recent work
4. **Diff Summary**: Run `git diff --stat` to see what's changed but not committed

Example status report format:
```
## Worktree Status: ml-feature-name

**Progress from worktree-claude:**
- [Summary from CLAUDE.md Progress section]

**Git Status:**
- infrastructure/: 3 files modified, ready to commit
- intercom/: Clean working directory, 2 commits ahead

**Recent Activity:**
- [Recent commits and changes]
```

### Working on Features

All work happens within the feature directory:

```bash
cd worktrees/ml-feature-name

# Make changes across repositories
cd intercom && git commit -m "Add new feature endpoint"
cd ../infrastructure && git commit -m "Add infrastructure for feature"
cd ../muster && git commit -m "Update deployment config"
```

### Publishing and Creating PRs

From within each worktree:

```bash
# In each relevant repo worktree
git publish              # Push branch with upstream tracking
pro                     # Create pull request
pru                     # Update PR from commit message
prfeed [reviewer]       # Post to PRFeed for review
```

### Cleanup After Merge

```bash
# Remove worktrees after feature is merged
cd ~/src
git worktree remove worktrees/ml-feature-name/intercom
git worktree remove worktrees/ml-feature-name/infrastructure
git worktree remove worktrees/ml-feature-name/muster
rm -rf worktrees/ml-feature-name
```

## Benefits

1. **Context Isolation**: Each feature has its own workspace with all relevant repos
2. **No Context Switching**: All related changes are visible in one directory
3. **Coordinated Development**: Easy to see relationships between changes across repos
4. **Clean History**: Feature branches are isolated per repository
5. **Parallel Work**: Can work on multiple features simultaneously

## Best Practices

### Naming Conventions
- Feature directories: `ml-descriptive-feature-name`
- Branch names: Match the feature directory name for consistency
- Use lowercase with hyphens for compatibility

### Repository Selection
- Only add worktrees for repos that will actually change
- Common combinations:
  - `intercom` + `infrastructure` for new features
  - `intercom` + `muster` for deployment changes
  - `infrastructure` + `event-management-system` for platform changes

### Git Workflow Integration
- Follow existing branch naming: `ml/feature-name` 
- Use standard commit workflow: commit → amend → publish → pro → pru → prfeed
- Handle rebases with `git resync` when needed
- Link PRs appropriately with `For: <issue-link>` (default) or `Closes: <issue-link>` (when explicitly closing)

## Troubleshooting

### Worktree Issues
```bash
# List all worktrees
git worktree list

# Prune stale worktree references
git worktree prune

# Force remove problematic worktree
git worktree remove --force path/to/worktree
```

### Common Scenarios

**Stale branch errors**: Use `git resync` in the affected worktree
**Permission issues**: Ensure worktree parent directories exist
**Cleanup failures**: Use `git worktree prune` then manual `rm -rf`

## Repository Coordination

This `~/src` directory itself is a Git repository that tracks:
- This coordination documentation (`CLAUDE.md`)
- Any workflow documentation (`*.md` files)
- The `.gitignore` that maintains this clean tracking

It explicitly ignores all repository contents while allowing documentation to be versioned and shared across the team.