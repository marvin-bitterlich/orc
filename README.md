# ORC - Orchestrator Command Center

**Forest Factory Command Center for El Presidente's Development Ecosystem**

ORC coordinates development workflow through universal commands, lightweight planning, and efficient worktree orchestration. One command system accessible everywhere, one planning approach that scales, one coordination layer that just works.

## What ORC Provides

**🎯 Universal Commands** - Access your development toolkit from any Claude Code session  
**📋 Lightweight Planning** - Simple 3-state planning without ceremony  
**🌳 Clean Worktrees** - Isolated development environments with automated setup

## Quick Examples

```bash
# Universal commands work everywhere
/tech-plan new-investigation     # Create focused tech plan
/bootstrap                       # Load project context  
/janitor                         # Maintain and cleanup

# Simple planning workflow
tech-plans/backlog/     → in-progress/     → archive/
   (future work)        (active projects)    (completed)

# Clean worktree setup
git worktree add ~/src/worktrees/ml-feature-repo -b ml/feature
cd ~/src/worktrees/ml-feature-repo
# Automated tech plan integration via symlinks
```

## How It Works

**Commands** (`global-commands/`) are symlinked globally for universal access  
**Plans** (`tech-plans/`) flow through backlog → in-progress → archive states  
**Worktrees** link to plans via symlinks for integrated development

## Documentation & Architecture

Complete technical documentation, architecture details, and implementation guides available in the `docs/` directory.

## Experimental Work

The `experimental/` directory contains prototypes and experimental systems, including an MCP task management server for potential future integration.

---

**Orchestrator Claude Coordinates. Investigation Claude Implements. El Presidente Commands.**