---
name: ship-plan
description: C2/C3 engineering review that pressure-tests synthesized knowledge and creates tasks. Use when ready to convert exploration/synthesis into actionable implementation work.
---

# Ship Plan

Transform synthesized knowledge into actionable tasks through a structured engineering review. Operates at C2 (containers) and C3 (components) zoom level.

## Usage

```
/ship-plan              (plan focused shipment)
/ship-plan SHIP-xxx     (plan specific shipment)
```

## When to Use

- After /ship-synthesize has produced a summary note
- When exploration is complete and ready to convert to tasks
- To pressure-test ideas against reality before implementation

## Documentation Discovery

Look for architecture and development docs in order:
1. `docs/reference/architecture.md` - System structure (C2/C3 mapping)
2. `CLAUDE.md` - Development rules and patterns

If found, reference during engineering review.
If not found, proceed with codebase exploration.

## Zoom Level (C4 Model)

| Level | Scope | Owned By |
|-------|-------|----------|
| C1: System Context | External systems, actors | ship-plan (identify) |
| C2: Container | Services, apps, databases | **ship-plan (primary)** |
| C3: Component | Modules within containers | **ship-plan (secondary)** |
| C4: Code | Files, functions, edits | IMP (implementation) |

**ship-plan** creates tasks at C2/C3: "touch the auth component in the CLI container"
**IMP** operates at C4: "edit cmd/auth.go lines 45-60"

## Flow

### Step 1: Get Shipment

If argument provided:
- Use specified SHIP-xxx

If no argument:
```bash
orc focus --show
```
- Use focused shipment
- If no focus: "No shipment focused. Run `orc focus SHIP-xxx` first."

### Step 2: Prerequisites Check

```bash
orc shipment show SHIP-xxx
orc note list --shipment SHIP-xxx
```

**Check for summary note:**
Look for a note with type=spec or title starting with "Summary:".

If no summary found:
```
⚠️ No summary note found in SHIP-xxx.

Consider running /ship-synthesize first if notes are messy.
Proceeding anyway...
```

**Check for development docs:**
Look for CLAUDE.md or docs/ for existing patterns.

If found:
```bash
# Review for existing patterns and checklists
```
Reference relevant sections during interview.

If not found, proceed with codebase exploration.

### Step 3: Gather Context

Read all open notes, especially:
- Summary note (if exists)
- Spec notes
- Decision notes

Build understanding of:
- What needs to be built
- What decisions have been made
- What constraints exist

---

## Engineering Review (5 Phases)

Each phase uses the orc-interview format:
- Natural language questions
- Options: 1=approve, 2=alternative, 3=skip, 4=discuss
- Max 5 questions per phase
- Progress indicator [Question X/Y]

### Phase 1: Environmental Assumptions

Surface and validate assumptions about the environment.

**What to check:**
- Dependencies that must exist
- Existing behavior being relied upon
- Infrastructure or tooling assumptions
- Permission or access assumptions

**Example questions:**
```
[Question 1/5]

The summary assumes the orc note close command supports --reason and --by flags.
This is fundamental to the synthesis flow. If it doesn't exist, we'd need to
add it first.

Based on our CLI exploration, this command does exist and works as expected.

1. Approve - assumption validated
2. Needs verification - let me check
3. Skip
4. Discuss
```

```
[Question 2/5]

The plan depends on skills being deployed automatically via git hooks. This
means changes to glue/skills/ will appear in ~/.claude/skills/ after commit.
If this isn't working, IMPs won't see new skills.

1. Approve - git hooks are configured
2. Need to verify hook setup
3. Skip
4. Discuss
```

### Phase 2: Specification Gaps

Identify underspecified areas that would force IMP decisions.

**What to check:**
- Edge cases not covered
- Behavior in error scenarios
- Format/structure details left vague
- Authority or ownership unclear

**Goal:** IMPs should implement without asking questions. Gaps requiring human judgment must be resolved here.

**Example questions:**
```
[Question 1/5]

The ship-synthesize flow says "identify themes" but doesn't specify what
happens if there's only one note. Should it skip theme identification and
go straight to summary, or still run through the full flow?

I'd recommend: skip theme identification with <3 notes, just summarize directly.

1. Approve - skip themes for small note counts
2. Always run full flow
3. Skip - let IMP decide
4. Discuss
```

**Resolution options:**
- **Specify now**: Add detail to summary note or task description
- **Defer to IMP**: Mark as acceptable ambiguity
- **Out of scope**: Explicitly exclude from this shipment

### Phase 3: System Touchpoints (C2/C3 Mapping)

Map what containers and components are affected.

**Reference:** If docs/reference/architecture.md exists, use it for C2/C3 structure. Otherwise, explore the codebase to identify containers and components.

**Example output:**
```
Systems affected (C2 - Containers):

| Container | Touched | Notes |
|-----------|---------|-------|
| Skills (glue/skills/) | ✓ | Create/modify skill files |
| CLI (cmd/, internal/) | ✗ | No CLI changes |
| Database | ✗ | No schema changes |
| Config | ✗ | No config changes |
| Documentation | ✓ | CLAUDE.md |

Components within Skills (C3):
- ship-synthesize/ [NEW]
- ship-plan/ [MODIFY]
- orc-interview/ [NEW]

Confirm scope? [y/edit/expand]
```

### Phase 4: Tooling Compatibility

Check against development tooling and workflows.

**Checklist:**
```
Tooling compatibility:

| Item | Status | Notes |
|------|--------|-------|
| Works with test setup? | ✓ | Skills are markdown, no code tests |
| New CLI commands needed? | ✓ | Uses existing orc commands |
| Database migrations? | ✓ | None required |
| Git hooks deploy? | ✓ | Skills auto-deploy |
| Config schema changes? | ✓ | None |
| Doc updates needed? | ⚠️ | CLAUDE.md |

Issues found: [list or none]
```

### Phase 5: Reality Checks

Final validation against known constraints.

**What to check:**
- Conflicts with existing behavior
- Technical limitations
- Resource constraints
- Timeline dependencies

**Example questions:**
```
[Question 1/5]

Deleting ship-tidy and exorcism is a breaking change for anyone using those
commands. They'll get "skill not found" errors. Since this is an internal
tool, that's probably fine - users can adapt. But worth confirming.

1. Approve - breaking change is acceptable
2. Add deprecation period
3. Skip
4. Discuss
```

---

## Task Generation

After engineering review, generate tasks.

### Task Format

Each task includes C2/C3 scope:

```
Proposed tasks for SHIP-xxx:

1. Create orc-interview skill
   Containers: Skills
   Components: glue/skills/orc-interview/
   Description: Foundation skill for structured interviews...

2. Create ship-synthesize skill
   Containers: Skills
   Components: glue/skills/ship-synthesize/
   Depends on: #1 (uses orc-interview as a primitive)
   Description: Knowledge compaction skill...

3. Update documentation
   Containers: Documentation
   Components: CLAUDE.md
   Description: Document new workflow...

Create these tasks? [y/n/edit]
```

### Dependency Analysis

Before creating tasks, analyze which tasks depend on others:

- **Sequential**: Tasks touching the same compilation units or where one consumes another's output must have `depends_on` set. Example: a skill that imports another skill being created in the same shipment.
- **Independent**: Tasks touching different directories or components with no shared state can run in parallel. Leave `depends_on` empty.
- **Tasks with no depends_on are implicitly parallelizable** by IMP workers.

When in doubt, prefer independence. Only add dependencies when one task genuinely cannot start until another completes.

### Task Creation

For each approved task:
```bash
orc task create "<Title>" \
  --shipment SHIP-xxx \
  --description "<Description with C2/C3 scope>"

# If task depends on others, specify dependencies:
orc task create "<Title>" \
  --shipment SHIP-xxx \
  --depends-on TASK-001 --depends-on TASK-002 \
  --description "<Description with C2/C3 scope>"
```

### Summary Output

```
Shipment planned:
  SHIP-xxx: <Title>
  Status: ready
  Tasks created: X

Engineering review complete:
  ✓ Assumptions validated: N
  ✓ Gaps resolved: N
  ✓ Systems mapped: N containers, M components
  ✓ Tooling compatible
  ✓ Reality checks passed

Ready for implementation:
  orc task list --shipment SHIP-xxx
```

## Guidelines

- **Check for development docs** (CLAUDE.md, docs/) for existing patterns
- **Tasks should be self-contained** - IMP can complete without questions
- **Include C2/C3 scope** in every task description
- **Set depends_on for sequential tasks** - tasks sharing compilation units or consuming each other's output
- **Leave depends_on empty for parallel tasks** - tasks touching different components
- **Don't over-decompose** - 3-10 tasks typical
- **Validate assumptions** - don't pass uncertainty to IMPs

## Task vs Plan Distinction

| Artifact | Created By | Zoom Level | Contains |
|----------|------------|------------|----------|
| Task | ship-plan | C2/C3 | What systems/components to touch |

Tasks are work packages with scope. IMP agents handle C4 implementation details.

## Example Session

```
> /ship-plan

[gets focused shipment SHIP-276]
[checks for summary note - found NOTE-514]
[reads CLAUDE.md]

Planning SHIP-276: Skill Cognitive Redesign

Prerequisites:
  ✓ Summary note found: NOTE-514
  ✓ CLAUDE.md reviewed

Starting engineering review...

## Phase 1: Environmental Assumptions

[Question 1/5]

The summary assumes skills in glue/skills/ are auto-deployed via git hooks...

1. Approve
> 1

[continues through all 5 phases...]

## Task Generation

Proposed tasks for SHIP-276:

1. Create orc-interview skill
   Containers: Skills
   Components: glue/skills/orc-interview/

2. Create ship-synthesize skill
   Containers: Skills
   Components: glue/skills/ship-synthesize/
   Depends on: #1 (uses orc-interview primitive)
   ...

Create these 8 tasks? [y/n/edit]
> y

[creates tasks...]

Shipment planned:
  SHIP-276: Skill Cognitive Redesign
  Status: ready
  Tasks created: 8

Ready for implementation:
  orc task list --shipment SHIP-276
```
