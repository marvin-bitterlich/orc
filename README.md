# 🏭 ORC 2.0 - Forest Factory Orchestration System

**Forest Factory Command Center for El Presidente's Development Ecosystem**

![Forest Factory](assets/orc.png)

## 🎯 What It Does

ORC 2.0 coordinates multiple Claude agents working across repositories simultaneously. 

## 🗂️ Two-Layer Memory Architecture

- **SQLite Ledger** - Structured execution tracking work in hierarchical missions → work orders → groves, with a mail-style handoff system for rapid context transfer between sessions
- **Graphiti Knowledge Graph** - Long-term intelligence that remembers design decisions, discoveries, and learnings across all investigations

## 🌲 Key Concepts

- **🎭 ORCs (Orchestrators)** - Coordinate missions, manage work orders, and facilitate cross-grove communication
- **⚒️ IMPs (Implementers)** - Work in groves on actual code, completing work orders and reporting discoveries back
- **🌳 Groves** - Git worktrees where IMPs do their work; one mission can have multiple groves from different repositories
- **🔮 Oracle** - Learns preferences and patterns over time, providing guidance to agents based on historical decisions and El Presidente's style

## 🤖 Systematic + Intelligent

Claude agents work autonomously on separate cross-repo tasks while sharing knowledge through the dual-layer memory. The system combines systematic execution (structured work tracking, TMux coordination) with self-reflection (semantic context, architectural memory). Currently powering all development work with full context preservation across sessions.

---

## Quick Start

```bash
# Initialize ORC context
orc prime

# View current mission status
orc status

# List work orders
orc work-order list

# Create a new mission
orc mission create "Mission Title" -d "Description"

# Create work orders
orc work-order create "Task title" --mission MISSION-001

# View mission summary
orc summary
```

## Command System Architecture

**Central Management + Global Access**
- Commands stored in `global-commands/` (universal) and `.claude/commands/` (ORC-specific)
- Symlinked to `~/.claude/commands/` for global availability
- Single source of truth - update once, available everywhere

**Key Commands**
- `/handoff` - Create handoff with automatic session restart
- `/g-bootstrap` - Full context restoration (ledger + Graphiti + disk)
- `orc prime` - Lightweight context injection
- `orc status --handoff` - View latest handoff

---

**Orchestrator Claude Coordinates. Investigation Claude Implements. El Presidente Commands.**
