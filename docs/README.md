# ORC Ecosystem Documentation

**Forest Factory Command Center - Architecture and Operations**

This documentation describes the complete ORC (Orchestrator) development ecosystem that provides centralized command management, lightweight tech planning, and efficient worktree coordination for El Presidente's development workflow.

## Overview

The ORC ecosystem solves three core problems:
1. **Command Discoverability** - Universal access to project management commands
2. **Cross-Repository Confusion** - Clean single-repo worktree architecture  
3. **Lightweight Planning** - Quick tech plan creation without ceremony

## Architecture Components

The ORC ecosystem consists of interconnected systems documented in this directory:

**Core Systems**:
- **Worktree System** - Single-repository worktrees with symlinked tech plans  
- **Tech Plans System** - Lightweight 3-state planning (backlog → in-progress → archive)
- **Command System** - Universal slash commands accessible globally via symlinks
- **Orchestrator Workflow** - Coordination procedures for worktree and TMux management

**Supporting Systems**:
- **Tools Evaluation** - Framework for assessing development tools and workflow enhancements
- **Integration Patterns** - How all components work together seamlessly

*See individual `.md` files in this directory for detailed implementation guidance.*

## Quick Reference

### Complete Directory Structure
```
orc/
├── docs/                    # Complete ecosystem documentation
├── global-commands/         # Universal command definitions (symlinked globally)
├── tech-plans/              # Central planning system
│   ├── in-progress/         # Active worktree investigations
│   ├── backlog/            # Future work items
│   └── archive/            # Completed work
├── experimental/            # Experimental systems and prototypes
│   └── mcp-server/         # Rails-based MCP task management system
├── .claude/
│   └── commands/           # ORC-specific command definitions  
├── work-trees -> ~/src/worktrees/  # Symlink to active worktrees
├── CLAUDE.md               # Claude session context
└── README.md               # Project introduction
```

### Key Symlinks
```
~/.claude/commands/ → orc/global-commands/      # Global command access
worktree/.tech-plans → orc/tech-plans/in-progress/[worktree]/  # Local tech plans
```

### Worktree States
```
~/src/worktrees/[active]/   # Currently working on
~/src/worktrees/paused/     # Valid but not active focus
[deleted]                   # Completed and archived
```

## Getting Started

1. **Create New Investigation**: Use orchestrator to set up single-repo worktree
2. **Plan Work**: Use `/tech-plan` command for lightweight planning
3. **Manage Progress**: Tech plans flow from backlog → in-progress → archive
4. **Pause/Resume**: Move worktrees between active and paused states
5. **Maintain System**: Use `/janitor` for cleanup and organization

## Current Implementation Status

### ✅ **Operational Systems**
- **Command System**: 9 universal commands fully operational
- **Symlink Architecture**: Global command access working across all contexts  
- **Tech Plans Structure**: 3-state organization (backlog/in-progress/archive) complete
- **Worktree Architecture**: Single-repo focus with symlinked tech plans validated
- **Documentation**: Complete ecosystem documentation and workflow procedures

### 🔄 **Integration & Evolution**
- **Command Modernization**: Some commands need updates for current architecture patterns
- **Legacy Migration**: Existing worktrees transitioning to new symlink pattern
- **Experimental Systems**: MCP task management server foundations built, purpose evolving

### 📋 **Development Priorities**
- Modernize command implementations for consistency with current architecture
- Complete migration of legacy worktree patterns to symlink-based approach
- Evaluate experimental systems for integration or archival decisions