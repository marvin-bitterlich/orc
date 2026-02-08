# ORC Glossary

**Purpose**: Centralized definitions for all ORC concepts, separated by system/concern

This directory contains the canonical definitions for ORC terminology. When there's confusion about what something means, start here.

---

## Files

### [ledger-entities.md](./ledger-entities.md)
**System**: SQLite database (~/.orc/orc.db)
**Contains**: Commission, Workshop, Workbench, Shipment, Task, Handoff
**Purpose**: Database schema definitions and entity relationships
**Status**: Current schema reference

### [forest-factory-roles.md](./forest-factory-roles.md)
**System**: Place-based actor model
**Contains**: El Presidente, IMP, Goblin, Watchdog, Workbench (as roles/places)
**Purpose**: Personality-driven language and role definitions
**Status**: Current role definitions

---

## Actor Model

ORC uses a **place-based actor model** where identity is tied to "where you are":

| Actor | Place | Role |
|-------|-------|------|
| IMP | Workbench (BENCH-XXX) | Implementation |
| Goblin | Gatehouse (GATE-XXX) | Review/Coordination |
| Watchdog | Watchdog (WATCH-XXX) | Monitoring |

Identity derives from the `place_id` in `.orc/config.json`.

---

## Shipment Workflow

The standard flow for turning exploration into implementation:

```
exploring (messy notes)
  → /ship-synthesize → Summary note
  → /ship-plan → Tasks (C2/C3 scope)
  → /imp-plan-create → Plans (C4 file detail)
  → Implementation
  → /imp-rec → Completion
```

---

## How to Use This Glossary

**When adding new concepts**:
1. Determine which system it belongs to (Ledger entities or Roles)
2. Add definition to appropriate file
3. Note any relationships with existing concepts

**When something is unclear**:
1. Check this glossary first
2. Check CLAUDE.md for operational details
3. Check docs/architecture.md for system overview

---

## Related Documents

- **CLAUDE.md** - Development rules and checklists
- **docs/architecture.md** - System architecture overview
- **README.md** - Technical overview

---

## Glossary Evolution

This glossary is **living documentation**:
- Add new concepts as they emerge
- Update after architectural decisions
- Remove deprecated concepts

**Version in git** to track evolution over time.

---

**Last Updated**: 2026-02-08
**Status**: Current reference documentation
