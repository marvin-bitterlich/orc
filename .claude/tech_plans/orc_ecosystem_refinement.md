# ORC Ecosystem Refinement

**Status**: implementation

## Problem & Solution
**Current Issue:** ORC worktree system has navigation confusion, tooling integration problems, and inconsistent command availability that reduces development velocity
**Solution:** Systematic refinement of worktree architecture, global command system, and enhanced context management to create seamless development experience

## Current System Analysis

### ✅ What's Working Well
- **Worktrees**: Excellent isolation and feature-based development
- **DLQ Bot Architecture**: 5-specialist sub-agent system operational  
- **TMux Integration**: Automated window setup with proper layouts
- **Development Workflows**: Clean separation between dev/test and production

### ❌ Core Pain Points Identified

#### **1. Cross-Repository Navigation Confusion**
- Claude gets confused about which repository it should work in within multi-repo worktrees
- Path confusion leading to errors and inefficient context switching
- Root cause: Worktree coordination directory vs individual git repository roots

#### **2. Vim/Fugitive Integration Breakdown**
- Fugitive requires vim root set to git directory, doesn't work from worktree root
- Cannot effectively view git diffs from worktree coordination level
- Broken git integration in primary code viewing tool
- **Potential Solution**: claude-code.nvim provides direct Neovim integration with git project root detection

#### **3. Command Discoverability Crisis**
- Valuable project management commands buried in individual repositories
- Duplicated effort across projects, inconsistent tooling experience
- intercom-bot-test had sophisticated automation hidden from ecosystem

#### **4. Context Management Gaps**
- No systematic approach to context management across worktrees
- Lost context when switching between projects
- Poor task handoff between orchestrator and implementation Claude sessions

## Implementation

### Approach
**Phase-based systematic refinement**: Address command discoverability first (immediate wins), then tackle navigation architecture, finally implement advanced context management.

### Current Progress
**✅ Phase 1 Complete**: Global command system operational
- `/janitor` - Complete project maintenance 
- `/tech-plan` - Structured technical planning
- `/analyze-prompt` - Advanced prompt quality assessment
- `/analyze-workflow` - Workflow analysis capabilities
- `/bootstrap` - Project bootstrapping automation

### Interface/API/Contract
**Global Command Architecture**:
```
~/.claude/commands/
├── janitor.md           → Project maintenance automation
├── tech-plan.md         → Structured planning system  
├── analyze-prompt.md    → AI prompt quality control
├── analyze-workflow.md  → Workflow analysis
└── bootstrap.md         → Project initialization
```

**Worktree Strategy Evolution**:
- Current: Multi-repo worktrees causing confusion
- Target: Single-repo worktrees with orchestrator-level coordination
- Vim Integration: Set working directory to git repository root by default

## Testing Strategy

### Phase 1: Global Commands (✅ Complete)
- [x] Command migration without disruption to existing workflows
- [x] Verification that commands work across all Claude sessions
- [x] Preservation of original commands in source repositories

### Phase 2: Single-Repo Worktree Testing
- [ ] Create test worktree with single repository
- [ ] Validate vim/fugitive integration works properly
- [ ] Measure navigation efficiency improvements
- [ ] Test cross-repository coordination at orchestrator level

### Phase 3: Context Management Integration
- [ ] Evaluate claude-task-master for systematic task tracking
- [ ] Test context handoff protocols between sessions
- [ ] Validate cross-worktree status reporting
- [ ] Measure context preservation across session boundaries

## Implementation Plan

### Phase 1: Command System Modernization ✅ COMPLETE
**Goal**: Extract and globalize useful project management commands
- [x] **Command Audit**: Found 5 sophisticated commands from intercom-bot-test
- [x] **ORC-Centric Architecture**: Moved master commands to `/Users/looneym/src/orc/.claude/commands/`
- [x] **Global Symlink System**: Created global symlinks (`~/.claude/commands/` → ORC masters)
- [x] **Command Validation**: Tested symlinked command functionality

**Current Commands**:
- `analyze-prompt.md` - Universal prompt quality analysis
- `bootstrap.md` - Universal project bootstrapping
- `janitor.md` - Universal project maintenance (currently running!)
- `tech-plan.md` - Universal technical planning

**Result**: ORC-controlled, universally accessible command system operational

### Phase 2: Global Architecture Consolidation 🔄 IN PROGRESS
**Goal**: Complete overhaul of global Claude ecosystem before adding new tools

**Current Progress**:
- [x] **Research**: Claude Code command organization and symlink capabilities
- [x] **ORC Command Architecture**: Symlink-based centralized command management
- [ ] **Global CLAUDE.md Overhaul**: Update to reflect new ORC-centric ecosystem
- [ ] **Global Agents Population**: Extract and organize universal agents
- [ ] **Local Integration**: Update worktrees with symlink patterns

#### **Current Global State Analysis**
- ✅ **Commands**: Recently migrated sophisticated command system
- ❌ **Agents**: Empty directory, needs population from local repositories  
- ❌ **CLAUDE.md**: Outdated, references old spellbook system, needs complete overhaul
- ❌ **Symlink Architecture**: No pinning system for easy vim/NerdTree navigation

#### **Symlink Pinning Strategy** (from bot test repo pattern)
**Current Pattern in Bot Test**:
```
.github/bots/dlq-bot/
├── agents -> ../../../.claude/agents     # Easy access to local agents
├── cache -> ../../../log/dlq-cache      # Easy access to cache/logs
├── investigations -> ../../../log/workflow-investigations  
└── workflow.yml -> ../../workflows/dlq-bot.yml
```

**Global Architecture Design**:
- **Actual storage**: Logical locations (~/.claude/commands, ~/.claude/agents, spellbooks)
- **Access convenience**: Symlinks in development contexts for NerdTree navigation
- **Responsibility split**: Clear delineation between global ecosystem and local project needs

#### **Responsibility Split Framework**

**Global (~/.claude/) - The Universal Foundation**:
- **Commands**: Universal project management (`/janitor`, `/tech-plan`, `/analyze-prompt`, etc.)
- **Core Agents**: General-purpose specialists (worktree analysis, prompt engineering, etc.)
- **CLAUDE.md**: Personality, relationship dynamics, ecosystem overview, worktree system documentation
- **Configuration**: MCP servers, global settings, universal patterns

**Local (.claude/ in repositories) - Project Specialization**:
- **Domain Agents**: Project-specific specialists (DLQ investigation, Marketo analysis, etc.)
- **Local Commands**: Project-specific automation and workflows
- **Context**: Project history, domain knowledge, specialized procedures
- **Symlinks**: Convenient access to relevant global resources for NerdTree navigation

**Symlink Strategy Pattern**:
```
project/.claude/
├── agents/                    # Local project agents
├── commands/                  # Project-specific commands  
├── global-agents -> ~/.claude/agents    # Easy access to universal agents
├── global-commands -> ~/.claude/commands # Quick reference to global commands
└── spellbook -> ~/src/orc/spellbook/    # Access to detailed procedures
```

#### **Global CLAUDE.md Architecture Redesign**

**Current Problems**:
- References outdated spellbook system in `orc/spellbook/`
- Missing documentation of worktree system
- No command discovery guidance
- Lacks ecosystem overview

**New Global CLAUDE.md Structure**:
```markdown
# CLAUDE.md - ORC Ecosystem Foundation

## El Presidente Relationship & Personality
[Current relationship dynamics - preserve existing tone]

## ORC Development Ecosystem Overview
### Worktree-Based Development
- Multi-repo feature isolation system
- Orchestrator vs Implementation Claude roles
- TMux integration and automated setup

### Global Command System  
- Universal project management automation
- Command discovery and usage patterns
- Integration with local project workflows

### Agent Architecture
- Global general-purpose agents
- Local project specialization pattern
- Sub-agent coordination protocols

## Quick Reference
### Essential Commands
- `/janitor` - Complete project maintenance
- `/tech-plan` - Structured technical planning
- `/analyze-prompt` - AI prompt quality control

### Navigation Patterns
- Symlink pinning for NerdTree access
- Global vs local resource organization
- Cross-worktree coordination protocols
```

#### **Global Agents Migration Strategy**

**Agents to Extract from Local Repositories**:
1. **Worktree Analysis Specialist** (from orc/spellbook) - Cross-worktree status reporting
2. **Prompt Engineering Analyst** (from bot test commands) - Universal prompt quality control
3. **Technical Planning Specialist** (from bot test commands) - Structured planning across projects
4. **Project Maintenance Agent** (from janitor command logic) - Universal cleanup automation

**Agent Organization**:
```
~/.claude/agents/
├── analysis/
│   ├── worktree-analysis-specialist.md
│   ├── prompt-engineering-analyst.md  
│   └── workflow-analyzer.md
├── planning/
│   ├── technical-planning-specialist.md
│   └── project-bootstrap-agent.md
└── maintenance/
    ├── project-janitor.md
    └── code-quality-auditor.md
```

#### **Implementation Sequence for Global Overhaul**

**Step 1: Global CLAUDE.md Overhaul**
- Preserve existing personality and relationship dynamics
- Add comprehensive ORC ecosystem documentation
- Document worktree system and orchestrator/implementation roles
- Create command discovery and navigation guidance
- Establish symlink architecture documentation

**Step 2: Global Agents Population**
- Extract worktree analysis specialist from orc/spellbook
- Convert command logic to standalone agents (janitor → project-maintenance agent)
- Create prompt engineering analyst from analyze-prompt command
- Organize agents by category (analysis, planning, maintenance)
- Test agent accessibility across all Claude sessions

**Step 3: Symlink Pinning System**
- Establish standard symlink patterns for NerdTree navigation
- Create convenience links in active worktrees
- Document symlink strategy for future projects
- Test vim/NerdTree navigation efficiency

**Step 4: Local Repository Integration**
- Update local .claude directories with symlinks to global resources
- Migrate project-specific agents to local directories
- Establish clear local vs global boundaries
- Create migration templates for future projects

**Migration Safety Protocols**:
- ✅ **No Deletion**: All existing content preserved during migration
- ✅ **Incremental Testing**: Test each component before proceeding
- ✅ **Rollback Strategy**: Original configurations remain accessible
- ✅ **Compatibility**: Ensure existing workflows continue to work

### Phase 3: Repository Navigation Optimization 🎯
**Goal**: Solve cross-repository confusion and vim integration issues (deferred until global architecture complete)

#### **Single-Repository Worktree Strategy**
- Default to single-repository worktrees for implementation work
- Use cross-repository coordination at orchestrator level only  
- Individual implementation Claudes work in focused single-repo environments

#### **Vim/Fugitive Integration Fix**
- Update `muxup` command to cd into primary repository directory
- Set vim working directory to git repository root by default
- Maintain worktree coordination through orchestrator, not implementation level

#### **Smart Context Management**
```
Context Management Strategy:
├── Orchestrator Level: Cross-worktree coordination and status
├── Implementation Level: Single-repo focus with clear context  
├── Command Integration: Global commands work regardless of location
└── Navigation Aids: Smart switching between repos when needed
```

### Phase 3: Ecosystem Integration & Enhancement 🚀
**Goal**: Evaluate and integrate best-in-class Claude Code ecosystem tools for comprehensive workflow enhancement

#### **Tool Evaluation Pipeline**
**Currently Evaluating**:
1. **claude-task-master** (in progress): Context management and task tracking system
2. **claude-code.nvim** (new): Direct Neovim integration with git root detection
3. **awesome-claude-code ecosystem** (new): Rich collection of workflow tools

**Evaluation Criteria**:
- Integration complexity with existing ORC system
- Impact on navigation efficiency and context management  
- Compatibility with worktree-based development
- Quality of documentation and community support

### Phase 4: Context Management & Task Tracking 🚀
**Goal**: Implement systematic context management across sessions

#### **Architecture Assessment**
**Primary Integration Candidates**:
- **claude-task-master**: Context management and task tracking system
- **claude-code.nvim**: Direct Neovim integration solving vim/fugitive issues
- **claudekit**: Auto-save checkpointing, code quality hooks, 20+ specialized subagents
- **awesome-claude-code ecosystem**: Rich collection of workflow enhancement tools

**Technical Integration Strategy**:
- Design context handoff protocols between orchestrator/implementation sessions
- Implement progress persistence across session boundaries  
- Create cross-worktree awareness system for orchestrator
- Evaluate vim/neovim plugin integration for seamless editor workflow

#### **Enhanced Status Reporting**
```
Status Reporting System:
├── Individual Worktree Status: Implementation Claude updates
├── Cross-Worktree Summary: Orchestrator provides global view
├── Task Tracking: Clear progress indicators and next steps
└── Context Preservation: Session-independent state management
```

## Success Metrics

### **Navigation Efficiency**
- Reduced path confusion and context switching errors
- Faster git diff and code review workflows  
- Seamless vim/fugitive integration

### **Command Discoverability** ✅
- All useful commands globally available via slash commands
- Consistent tooling experience across all projects
- Reduced duplication and improved reusability

### **Context Management**
- Clear task handoff between orchestrator and implementation sessions
- Persistent context across session boundaries
- Effective cross-worktree progress tracking

### **Development Velocity**
- Faster project setup and navigation
- Reduced friction in common development tasks
- Enhanced automation and workflow integration

## Validation Strategy

### **Global Architecture Validation**
- [ ] **Command Consistency**: All global commands work identically across orchestrator and implementation sessions
- [ ] **Agent Accessibility**: Universal agents available from any Claude session without path confusion
- [ ] **CLAUDE.md Completeness**: New global documentation provides comprehensive ecosystem guidance
- [ ] **Symlink Navigation**: NerdTree navigation works seamlessly with pinned global resources

### **Local Integration Validation**
- [ ] **Project Specialization**: Local agents and commands complement (don't conflict with) global resources
- [ ] **Symlink Integrity**: Local symlinks to global resources remain functional across filesystem changes
- [ ] **Workflow Continuity**: Existing project workflows continue working without disruption
- [ ] **Migration Completeness**: All valuable local resources properly categorized as global vs local

### **System Integration Testing**
- [ ] **Cross-Worktree Coordination**: Orchestrator Claude can manage multiple implementation sessions effectively
- [ ] **Context Preservation**: Task handoff between orchestrator and implementation works reliably
- [ ] **Tool Interoperability**: Global commands and agents work correctly in all worktree configurations
- [ ] **Performance Impact**: No significant slowdown in common development operations

### **Success Criteria**
**Immediate Wins (Post-Phase 2)**:
- New Claude session can discover and use all project management tools within 30 seconds
- NerdTree navigation to global resources requires maximum 2 clicks from any project
- Universal agents provide consistent quality across all projects

**Long-term Impact (Post-Phase 3)**:
- 50% reduction in "Where is that command?" or "How do I..." questions
- Consistent project maintenance and planning patterns across entire ecosystem
- Seamless integration of new tools without disrupting existing workflows