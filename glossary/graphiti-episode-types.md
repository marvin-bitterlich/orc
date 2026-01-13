# Graphiti Episode Types

**Source**: Original GLOSSARY.md
**Purpose**: Standardized episode types for Graphiti knowledge capture
**Approach**: Name-based convention (prefix episode names with type)
**Status**: Living document - evolve as patterns emerge

---

## Episode Types

### Design Decision
**What**: Architectural and design choices with rationale
**When**: After brainstorming reaches conclusion, when we decide "we're doing X because Y"
**Naming**: `Design Decision: [What We Decided]`

**Examples**:
- "Design Decision: Human-Driven Work Order Transition (Three-Level Operational Model)"
- "Design Decision: Three-Layer Tool Usage Model for ORC/Graphiti"
- "Design Decision: Complementary Tool Stack for ORC"

**Key Elements to Capture**:
- What was decided
- Why we decided it (rationale)
- What alternatives were considered
- Trade-offs and implications

---

### Learning Artifact
**What**: Research outputs from investigating external systems, patterns, or capabilities
**When**: After investigating how something works (tools, patterns, technologies)
**Naming**: `Learning Artifact: [What We Learned About]`

**Examples**:
- "Learning Artifact: Beads CLI Initialization & Context Loading"
- "Learning Artifact: Neo4j Deterministic Query Capabilities"
- "Learning Artifact: Hybrid DB Architectures (Industry Patterns 2024-2026)"

**Subtypes** (optional specificity):
- `Learning Artifact: [Tool Name] - [Focus Area]` (e.g., "Beads CLI - Session State")
- `Learning Artifact: [Technology] Capabilities for [Use Case]`
- `Learning Artifact: Industry Patterns for [Domain]`

**Key Elements to Capture**:
- What we investigated
- Key findings
- How it works
- Relevance to ORC
- Useful patterns or anti-patterns discovered

---

### Investigation Report
**What**: Comprehensive analysis leading to definitive conclusions (usually GO/NO-GO decisions)
**When**: After exhaustive investigation of whether to adopt/integrate something
**Naming**: `Investigation Report: [What Was Investigated] - [Conclusion]`

**Examples**:
- "Investigation Report: TaskMaster - Architectural Incompatibility"
- "Investigation Report: ORC Task Management Alternatives - Enhancement Path"

**Difference from Learning Artifact**: Investigation Reports make DECISIONS based on research
**Key Elements to Capture**:
- What was investigated and why
- Comprehensive findings
- Decision reached (adopt/reject/defer)
- Rationale for decision
- Re-evaluation criteria (if applicable)

---

### Architectural Vision
**What**: North star documents defining system identity and long-term direction
**When**: Defining foundational strategy, restoring/evolving core vision
**Naming**: `Architectural Vision: [System/Component Name]`

**Examples**:
- "Architectural Vision: The Forest Factory (North Star v1.0.0)"
- "Architectural Vision: Cognitive Brain System"

**Key Elements to Capture**:
- Core vision statement
- System philosophy and principles
- Long-term goals
- What makes this system distinctive
- Version history (for living documents)

---

### Implementation Record
**What**: "We built X, here's what happened"
**When**: After completing implementation work, documenting what was built and lessons learned
**Naming**: `Implementation Record: [What Was Built]`

**Examples**:
- "Implementation Record: /g-bootstrap and /g-handoff Commands"
- "Implementation Record: ORC Template System Consolidation"

**Key Elements to Capture**:
- What was built
- Key implementation decisions
- Challenges encountered
- What worked well
- Lessons learned
- Status/completion

---

### Tech Plan
**What**: Structured project plans with phases, tasks, and implementation approach
**When**: Planning implementation of features or systems
**Naming**: `Tech Plan: [Project Name]`

**Examples**:
- "Tech Plan: Graphiti Memory Integration"
- "Tech Plan: Cognitive Brain System"

**Key Elements to Capture**:
- Problem being solved
- Proposed solution
- Implementation phases
- Dependencies and prerequisites
- Success metrics
- Status tracking

---

### Session Summary
**What**: Work session context for continuity (from /g-handoff)
**When**: Automatic on /g-handoff at session boundaries
**Naming**: `Session Summary: [Worktree/Topic] - [Date]`

**Examples**:
- "Session Summary: ml-auth-refactor - 2026-01-13"
- "Session Summary: ORC Orchestrator - 2026-01-13"

**Key Elements to Capture**:
- Session focus
- TODOs and their status
- Decisions made during session
- Discoveries
- Open questions
- Next steps

---

### Discovery
**What**: Real-time insight - "Found that X works this way"
**When**: During work, when discovering how systems actually behave
**Naming**: `Discovery: [What Was Found]`

**Examples**:
- "Discovery: Redis pub/sub pattern in worker system"
- "Discovery: Auth flow uses session tokens not JWT"
- "Discovery: System uses pattern X for Y"

**Guideline**: If it's in git history/code, probably not brain-worthy
**Good for Graphiti**: System-level patterns, architectural insights, cross-component relationships
**Not for Graphiti**: "Function X calls function Y" level details

**Note**: Start with this, evolve if volume becomes overwhelming

---

### Development Principle
**What**: Reusable patterns, standards, and guidelines for how we build tools and systems
**When**: After establishing a successful pattern that should be followed in future work
**Naming**: `Development Principle: [Pattern Name]`

**Examples**:
- "Development Principle: System-Level Go Tool Installation Pattern"
- "Development Principle: Git Hook Auto-Deployment"
- "Development Principle: CLI Command Design Standards"

**Key Elements to Capture**:
- The principle/pattern being established
- Why this should be the standard approach
- What problem it solves
- How to apply it (implementation steps)
- Examples of successful application
- When to deviate (exceptions)

**Difference from Design Decision**: Principles are reusable standards for future work, not one-time choices

---

## Query Patterns

### Finding Specific Types
```python
# Get all episodes in ORC group
episodes = get_episodes(group_ids=["orc"], max_episodes=50)

# Filter by type (name prefix)
design_decisions = [e for e in episodes if e.name.startswith("Design Decision:")]
learning_artifacts = [e for e in episodes if e.name.startswith("Learning Artifact:")]
```

### Semantic Search Within Type
```python
# Find learning artifacts about databases
search_nodes("database learning artifact", group_ids=["orc"])

# Find design decisions about architecture
search_memory_facts("architecture design decision", group_ids=["orc"])
```

---

## Guidelines for Capture

### When to Create an Episode

**DO create when**:
- ✅ Information valuable beyond current session
- ✅ Future sessions need this context
- ✅ Cross-investigation insight
- ✅ Decision with rationale
- ✅ Research findings from external systems

**DON'T create when**:
- ❌ Information already in git/code
- ❌ Temporary debug output
- ❌ Ephemeral implementation details
- ❌ Raw code snippets (use git)

### Choosing the Right Type

**Decision tree**:
1. Did we DECIDE something? → **Design Decision**
2. Did we ESTABLISH a reusable pattern/standard? → **Development Principle**
3. Did we RESEARCH something external? → **Learning Artifact**
4. Did we INVESTIGATE for GO/NO-GO? → **Investigation Report**
5. Did we BUILD something? → **Implementation Record**
6. Are we PLANNING to build? → **Tech Plan**
7. Did we DISCOVER a pattern during work? → **Discovery**
8. Is this a FOUNDATIONAL vision? → **Architectural Vision**
9. Is this SESSION context? → **Session Summary** (automated)

When in doubt, use **Discovery** for small things, **Learning Artifact** for research.

---

## Evolution Strategy

**This glossary will evolve**:
- Start with these types
- If Discovery becomes too granular → refine guidelines
- If new patterns emerge → add new types
- If types aren't used → remove them
- Version this file in git to track evolution

**Review cadence**: Quarterly or when friction emerges

---

## Integration with ORC Workflows

### During ORC Sessions (Planning/Brainstorming)
Primary types: **Design Decision**, **Learning Artifact**, **Investigation Report**, **Architectural Vision**

### During IMP Execution (Grove Work)
Primary types: **Discovery**, **Session Summary**, **Implementation Record**

### Cross-Level
**Tech Plans** can originate in ORC sessions, track progress during IMP execution

---

**Last Updated**: 2026-01-13
**Status**: Migrated from GLOSSARY.md
