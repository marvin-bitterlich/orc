# ORC Glossary

**Status**: Living document
**Last Updated**: 2026-02-09

A-Z definitions of ORC terminology. For schema details see [schema.md](../reference/schema.md). For lifecycle states see [shipment-lifecycle.md](../reference/shipment-lifecycle.md).

---

## Terms

**📋 Commission**
A body of work being tracked. Top-level organizational unit. Contains shipments.

**👔 El Presidente**
The human. Strategic decision maker and boss. Commands the forest.

**🏭 Factory**
A collection of workshops, typically representing a codebase or project area.

**👺 Goblin**
Coordinator agent. The human's long-running workbench pane. Creates/manages ORC tasks with the human. Memory and policy layer (what and why).

**👹 IMP**
Disposable worker agent spawned by Claude Teams. Executes tasks using Teams primitives. Execution layer (how and who).

**📝 Note**
Captured thought within a shipment. Types: idea, question, finding, decision, concern, spec.

**📐 Plan**
C4-level implementation detail. Specifies files and functions to edit.

**📦 Shipment**
Unit of work with a 4-status lifecycle: draft, ready, in-progress, closed. Contains tasks and notes.

**✔️ Task**
Specific implementation work within a shipment. Lifecycle: open, in-progress, closed (+blocked lateral state).

**📖 Tome**
Knowledge container at commission level. Holds notes for long-running reference.

**🔧 Workbench**
Git worktree where agents work. Isolated development environment with dedicated tmux window.

**🛠️ Workshop**
Collection of workbenches for coordinated work.
