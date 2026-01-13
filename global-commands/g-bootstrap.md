# G-Bootstrap Command

Restore session context from Graphiti's persistent memory and synthesize with current project state for seamless work continuation.

## Role

You are a **Session Context Restorer** that reconstructs working context from Graphiti's temporal knowledge graph and current project state, enabling Claude to resume work exactly where the previous session left off.

## Usage

```
/g-bootstrap [--worktree worktree-name] [--full]
```

**Purpose**: Prime new Claude session with:
- **Agent memory** from Graphiti (decisions, discoveries, TODOs)
- **Project context** from disk (git history, CLAUDE.md)
- **Synthesized briefing** combining memory + current state
- **Cross-investigation insights** (with --full flag)

**Perfect Companion to /g-handoff**: Restores context that was flushed in previous session.

## Process

<step number="1" name="detect_context">
**Detect Current Context:**
- Determine current working directory
- Detect worktree from path (~/src/worktrees/NAME)
- Check if in ORC orchestrator context
- Use --worktree flag if provided (override detection)

**group_id Priority (same as /g-handoff):**
1. --worktree flag (explicit override)
2. Auto-detect from ~/src/worktrees/[name] → "worktree-[name]"
3. ~/src/orc → "orc"
4. "unknown-session" (fallback)

**Check ORC Ledger:**
- Verify `orc` binary is available
- Read ~/.orc/metadata.json for latest handoff pointer
- Check if ledger database exists
</step>

<step number="2" name="read_ledger_handoff">
**Read Ledger Handoff (PRIORITY: Do This First):**

**Read Metadata Pointer:**
```bash
cat ~/.orc/metadata.json
```

Extract `current_handoff_id` and active context IDs.

**Load Handoff from Ledger:**
```bash
orc handoff show [handoff-id]
```

**Parse Handoff:**
- Read narrative note (Claude-to-Claude message)
- Extract active mission/operation/work-order/expedition
- Load todos snapshot if present
- Note Graphiti episode UUID if linked

**Display Immediately:**
```markdown
# 🚀 Ledger Bootstrap - [Context]

## 📝 **Handoff from Previous Claude** (HO-XXX)
[Full narrative note from previous session]

**Active Context:**
- Mission: [MISSION-ID] - [Title]
- Operation: [OP-ID] - [Title]
- Work Order: [WO-ID] - [Title]
```

**Benefits:**
- **Instant context** (<1 second, no waiting)
- **Narrative clarity** (Claude-to-Claude communication)
- **Structured relationships** (database queries available)

**If no handoff found:**
- Display: "🆕 Fresh start - no previous handoff found in ledger"
- Proceed with Graphiti and disk context only
</step>

<step number="3" name="query_graphiti_memory">
**Query Graphiti for Semantic Memory (ASYNC: After Ledger):**

**After displaying ledger handoff**, query Graphiti for deeper insights.

**Use MCP tools to retrieve context:**

1. **Get Recent Episodes** (most recent handoffs):
   ```
   mcp__graphiti__get_episodes(
     group_ids=[detected_group_id],
     max_episodes=5
   )
   ```
   Parse episodes for:
   - Last session timestamp
   - TODO state from previous session
   - Decisions and rationale
   - Technical discoveries
   - Open questions and blockers

2. **Search for Recent Work** (semantic query):
   ```
   mcp__graphiti__search_memory_facts(
     query="recent work on [worktree topic]",
     group_ids=[detected_group_id],
     max_facts=10
   )
   ```
   Extract:
   - Related work relationships
   - Cross-component discoveries
   - Temporal evolution of understanding

3. **Find Relevant Entities** (architecture/discoveries):
   ```
   mcp__graphiti__search_nodes(
     query="[worktree topic] architecture technical discoveries",
     group_ids=[detected_group_id],
     max_nodes=10
   )
   ```
   Identify:
   - Key components/modules
   - Architectural patterns discovered
   - System relationships understood

**If --full flag provided:**
- Query across ALL group_ids for cross-investigation insights
- Find related work in other worktrees
- Surface relevant patterns from past investigations

**Important:** This step happens AFTER ledger handoff is displayed. User already has context and can start working!
</step>

<step number="4" name="load_disk_context">
**Load Traditional Project Context (Existing Bootstrap Pattern):**

1. **Read CLAUDE.md** for project context:
   - Project purpose and repository structure
   - Development workflows and commands
   - Key tools and integrations

2. **Check Recent Git Activity:**
   - Last 5-7 commits to understand recent work
   - Current branch status
   - Uncommitted changes or work in progress
</step>

<step number="5" name="synthesize_hybrid_context">
**Synthesize Hybrid Briefing (Ledger + Graphiti + Disk):**

Merge all three context sources to create comprehensive briefing:

**Three-Tier Context:**
1. **Ledger** (instant, structured): Active work, narrative note
2. **Graphiti** (semantic, temporal): Discoveries, decisions, cross-investigation insights
3. **Disk** (current state): Git history, uncommitted changes, project docs

**Prioritization Logic:**
1. **Most urgent**: Current work from ledger handoff
2. **High value**: Open questions from handoff + new information
3. **Contextual**: Graphiti discoveries + git activity

**Cross-reference patterns:**
- Ledger active work + Git recent commits
- Handoff narrative + Graphiti semantic insights
- Open questions + new information available

Create integrated narrative showing:
- What was being worked on (from ledger)
- What changed since then (from git/disk)
- Deeper insights available (from Graphiti)
- What's ready to resume (synthesized)
</step>

<step number="6" name="generate_briefing">
**Generate Comprehensive Bootstrap Briefing:**

Display structured briefing combining all context sources.
</step>

## Briefing Template

```markdown
# 🚀 Hybrid Bootstrap - [Worktree Name]

## 📝 **Ledger Handoff** (from Previous Claude - HO-XXX)

**Created**: [timestamp]
**Active Context:**
- Mission: [MISSION-ID] - [Title]
- Operation: [OP-ID] - [Title]
- Work Order: [WO-ID] - [Title]

**Handoff Note:**
[Full narrative from previous Claude - displayed immediately]

---

## 🧠 **Semantic Memory** (from Graphiti)

**Key Decisions Made**:
- **[Decision]**: [Rationale]
  - Context: [When and why this was decided]

**Technical Discoveries**:
- **[Discovery]**: [What was learned]
  - Impact: [How this affects approach]

**Open Questions** (need attention):
- ❓ [Question] - Priority: [high/medium/low]

---

## 🎯 **Resume Points** (Synthesized from Ledger + Graphiti + Disk)

**PRIORITY 1: Continue In-Progress Work**
[Most urgent TODO from last session + current context]

**PRIORITY 2: Address Open Questions**
[Questions from last session + new information available]


---

## 🔗 **Cross-Investigation Insights** (if --full)
[Related work from other worktrees via semantic search]
[Patterns discovered in past investigations applicable here]

```

## Implementation Logic

**Context Synthesis Algorithm:**
```
function synthesizeContext():
  graphiti_context = queryGraphiti(group_id)
  disk_context = loadDiskContext()

  // Merge and prioritize
  resume_points = []

  // Priority 1: In-progress from last session
  for todo in graphiti_context.todos:
    if todo.status == "in_progress":
      resume_points.append(todo with current git context)

  // Priority 2: Open questions
  for question in graphiti_context.open_questions:
    if hasNewInformation(question, disk_context):
      resume_points.append(question with resolution path)
```

**Graceful Degradation:**
```
try:
  graphiti_context = queryGraphiti()
except GraphitiUnavailable:
  display("⚠️  Graphiti unavailable, falling back to disk-only bootstrap")
  return traditionalBootstrap()
```


## Advanced Features

**Full Cross-Investigation Query:**
```
/g-bootstrap --full
```
- Queries across ALL group_ids
- Surfaces related work from other investigations
- Shows patterns discovered elsewhere
- Useful for: finding similar problems, reusing solutions

**Explicit Worktree:**
```
/g-bootstrap --worktree ml-auth-refactor
```
- Bootstrap specific worktree context
- Useful when switching between investigations

**Fresh Start Detection:**
- If no Graphiti episodes found:
  - Display: "🆕 Fresh start - no previous session found"
  - Fall back to traditional disk-only bootstrap
  - Still valuable for new investigations

**Stale Session Handling:**
- If last session > 7 days ago:
  - Display: "⏰ Last session was [X] days ago"
  - Highlight what changed since then (git activity)
  - Prompt to review context before resuming

## Integration Notes

**Replaces /compact:**
- Traditional /compact: Lossy summarization
- /g-bootstrap: Lossless context restoration from Graphiti
- No need to summarize conversations anymore

**Enhances /bootstrap:**
- Traditional /bootstrap loads disk context only
- /g-bootstrap adds agent memory layer
- Hybrid approach: best of both worlds

**Works With:**
- /g-handoff: The flush counterpart
- TodoWrite: Reconstructs TODO state
- EnterPlanMode: Informs planning with discoveries

**Storage Independence:**
- All memory in Neo4j (no files in repos)
- Worktree switches seamless (memory follows you)
- Global dev environment memory store
- No coordination between multiple databases

## Error Handling

**Graphiti Unavailable:**
```
⚠️  Cannot connect to Graphiti (http://localhost:8001/mcp)
ℹ️  Falling back to traditional disk-only bootstrap
💡 Start Graphiti: cd ~/src/graphiti/mcp_server && docker compose up
```

**No Previous Sessions:**
```
🆕 Fresh start - no previous session found in Graphiti
ℹ️  Proceeding with traditional disk-only bootstrap
💡 Use /g-handoff at end of session to capture context
```

**MCP Tools Not Loaded:**
```
⚠️  Graphiti MCP tools not available
ℹ️  Check MCP configuration in ~/.claude.json
ℹ️  Falling back to traditional disk-only bootstrap
```

**Empty Worktree Detection:**
```
❓ Could not detect worktree from current directory
ℹ️  Using group_id: "unknown-session"
💡 Use --worktree flag to specify explicitly
```
