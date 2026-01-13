# ORC Glossary

**Purpose**: Centralized definitions for all ORC concepts, separated by system/concern

This directory contains the canonical definitions for ORC terminology. When there's confusion about what something means, start here.

---

## Files

### [ledger-entities.md](./ledger-entities.md)
**System**: SQLite database (~/.orc/orc.db)
**Contains**: Mission, Operation, Work Order, Expedition, Grove, Plan, Handoff, Dependency
**Purpose**: Database schema definitions and entity relationships
**Status**: ⚠️ Has many contentious questions about relationships

### [graphiti-episode-types.md](./graphiti-episode-types.md)
**System**: Graphiti + Neo4j semantic memory
**Contains**: Design Decision, Learning Artifact, Investigation Report, Session Summary, etc.
**Purpose**: Standardized episode naming conventions for knowledge capture
**Status**: ✅ Stable, migrated from original GLOSSARY.md

### [forest-factory-roles.md](./forest-factory-roles.md)
**System**: Conceptual model from NORTH_STAR.md
**Contains**: El Presidente, ORC, IMP, Mage, Grove (as roles/metaphors)
**Purpose**: Personality-driven language and role definitions
**Status**: ⚠️ Has questions about IMP as entity vs role

---

## Key Tensions to Resolve

### 1. **Expedition vs Work Order**
- **Ledger**: Expedition has optional `work_order_id`, can exist independently
- **Question**: What IS an expedition? Is it "work order in execution"?
- **Impact**: Core to understanding when/how to create entities

### 2. **IMP: Role vs Entity**
- **Ledger**: `assigned_imp` is TEXT field (e.g., "IMP-ZSH")
- **NORTH_STAR**: IMPs are specialized workers with guilds
- **Question**: Should IMPs be their own database table?
- **Impact**: Affects how we track specializations and capabilities

### 3. **Grove Lifecycle**
- **Ledger**: Grove has `expedition_id` foreign key
- **Question**: When are groves created? Can they exist without expeditions?
- **Impact**: Affects worktree creation workflow

### 4. **Plan vs Tech Plan**
- **Ledger**: `plans` table links to expeditions
- **Filesystem**: `tech-plans/` directory with markdown files
- **Question**: What's the relationship? Should table be renamed?
- **Impact**: How we integrate tech planning with ledger

### 5. **Work Order State Management**
- **NORTH_STAR**: Directory-based (work-orders/01-backlog/, etc.)
- **Ledger**: Database status fields ('backlog', 'in_progress', etc.)
- **Question**: Which is source of truth? Both? Synchronized?
- **Impact**: State transition workflow

---

## How to Use This Glossary

**When adding new concepts**:
1. Determine which system it belongs to (Ledger, Graphiti, or Conceptual)
2. Add definition to appropriate file
3. Note any relationships or tensions with existing concepts
4. Flag contentious questions for discussion

**When resolving tensions**:
1. Discussion with El Presidente
2. Document decision in appropriate file
3. Update database schema if needed
4. Capture rationale in Graphiti as "Design Decision" episode

**When something is unclear**:
1. Check this glossary first
2. If not defined, add it with "??? NEEDS CLARIFICATION ???"
3. Bring to ORC session for discussion

---

## Related Documents

- **NORTH_STAR.md** - Overall vision and philosophy
- **CLAUDE.md** - Working instructions for ORC orchestrator
- **README.md** - Technical overview
- **ARCHITECTURE.md** - *(To be created)* Entity relationships and lifecycle

---

## Glossary Evolution

This glossary is **living documentation**:
- ✅ Add new concepts as they emerge
- ✅ Flag tensions and questions immediately
- ✅ Update after architectural decisions
- ✅ Remove deprecated concepts

**Version in git** to track evolution over time.

---

**Last Updated**: 2026-01-13
**Status**: Initial structure - ready for discussion with El Presidente
**Next Step**: Resolve contentious questions in WO-004 discussion
