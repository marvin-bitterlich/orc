# 🏭 ORC - The Forest Factory

![Forest Factory](assets/orc.png)

Deep in the forest stands a factory. The ORC oversees operations from the command center while IMPs work in scattered workbenches, hammering out code. Shipments move through the system - bundles of tasks ready for delivery. Tomes accumulate knowledge. Handoffs pass the torch between shifts.

ORC is a CLI for structured AI-assisted development. It tracks commissions, organizes work into containers, preserves context across sessions, and provisions isolated workspaces. The forest runs on SQLite and git worktrees.

Today, ORC orchestrates a single agent working thoughtfully through well-planned tasks. Tomorrow, a Shipment will spawn a swarm of IMPs working in parallel. But swarms need solid foundations - planning, context preservation, merge strategies, quality controls. The factory is being built to scale.

## 🚀 Getting Started

```bash
git clone <repo-url>
cd orc
make bootstrap
```

Then run `/orc-first-run` in Claude Code for an interactive walkthrough.

## 🎭 The Cast

**🧌 The ORC** is your Orchestrator - the agent who oversees the forest, coordinates commissions, and maintains the big picture. The ORC doesn't write code directly; it manages the work, tracks progress, and ensures context flows between sessions.

**👹 IMPs** are Implementation agents. These mischievous workers inhabit workbenches and do the actual coding. Each IMP works in isolation, focused on its assigned tasks, reporting discoveries back to the Orchestrator.

**🌳 Workbenches** are where the work happens. Technically they're git worktrees - isolated copies of repositories where an IMP can make changes without affecting the main codebase. One commission might have several workbenches, each focused on different aspects of the work.

**🎯 Commissions** are the grand undertakings. Every piece of work belongs to a commission, giving it context and purpose.

### 📦 Containers

Work in ORC is organized into containers that hold related items:

**🚢 Shipments** (SHIP-*) are bundles of tasks ready for delivery - the primary unit of work that moves through the system.

**🏛️ Conclaves** (CON-*) are gatherings where discussions happen and decisions are made.

**📜 Tomes** (TOME-*) are books of accumulated knowledge - documentation that persists and grows.

### 🍃 Leaves

Inside containers live individual items: **Tasks** are deeds to be done. **Questions** are riddles awaiting answers. **Plans** are maps of intent. **Notes** are scattered thoughts worth preserving.

### 🔮 Rituals

**Handoffs** pass the torch between sessions. When one Claude session ends and another begins, the handoff narrative carries context forward - what was accomplished, what remains, what pitfalls to avoid.

**Priming** awakens context at session start. Run `orc prime` and the current commission, focus, and recent history flow into the conversation.

---

*🌲 The forest hums with industry. Shipments move through workbenches. IMPs hammer at their tasks. And the ORC watches over all. 🌲*
