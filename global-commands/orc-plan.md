# ORC Plan Command

Implementation planning ceremony - designs **how** to build what the CWO specifies.

## Overview

This command guides you through the structured planning ceremony that ensures:
- Plans are aligned with AGENTS.md architecture rules
- Plans are scoped to the active Cycle Work Order (CWO)
- Plans are captured in the ORC ledger for tracking
- El Presidente explicitly approves before implementation begins

**Boundary:** This skill is about **planning implementation**, not scoping work. The CWO (created by `/orc-cycle`) defines *what* to build. This skill plans *how* to build it.

**Workflow position:**
```
/orc-cycle → CWO created → El Presidente approves → /orc-plan → implement → CREC → /orc-cycle
```

**Prerequisite:** A CWO must exist and be approved before running this skill. If no CWO exists, run `/orc-cycle` first.

---

## PHASE 1: Context Gathering (MANDATORY)

You MUST complete all steps before entering plan mode.

### Step 1.1: Get Focused Shipment

```bash
./orc status
```

Note the focused **Shipment ID** (e.g., SHIP-207) and **Commission ID** (e.g., COMM-001).

### Step 1.2: Read AGENTS.md (NON-NEGOTIABLE)

```bash
cat AGENTS.md
```

**You MUST read and internalize AGENTS.md before proceeding.** This contains:
- Architecture rules (hexagonal/ports & adapters)
- Layer boundaries (core → ports → app → adapters → cli)
- Testing requirements (FSM-first, table-driven tests)
- Verification discipline (must run tests and lint)
- Checklists for common operations

**Your plan MUST follow these rules. No exceptions.**

### Step 1.3: Get Work Order

```bash
./orc work_order list --shipment-id SHIP-XXX
./orc work_order show WO-XXX
```

The Work Order defines the **overall outcome** and **acceptance criteria** for the shipment. Your plan contributes to this larger goal.

### Step 1.4: Get Active Cycle Work Order

```bash
./orc cwo list --shipment-id SHIP-XXX
./orc cwo show CWO-XXX
```

**Check:** The CWO must exist and be approved (Cycle status: `approved`). If no CWO exists or it's still in draft, stop and run `/orc-cycle` first to create/approve one.

The CWO defines **this cycle's specific scope**. Your plan must:
- Achieve the CWO outcome
- Meet all CWO acceptance criteria
- Stay within the CWO scope (don't gold-plate)

### Step 1.5: Note the Cycle ID

The CWO belongs to a Cycle (e.g., CYC-003). Note this ID - you'll need it when capturing the plan to the ledger.

---

## PHASE 2: Planning

### Step 2.1: Enter Plan Mode

Use the `EnterPlanMode` tool now.

### Step 2.2: Handle Stale Plan Files

**IMPORTANT:** If Claude Code loads a plan file from a previous session, DISCARD its contents entirely. Start fresh based on the CWO you just read.

### Step 2.3: Design Your Implementation Plan

Your plan must include:

1. **Goal Statement** - What the CWO outcome is (copy from CWO)

2. **Approach** - High-level strategy for achieving the goal

3. **Implementation Steps** - Concrete steps, referencing:
   - Specific files to create/modify
   - Which AGENTS.md checklist applies (e.g., "Add CLI Command", "Add Field to Entity")
   - Architecture layer each change belongs to

4. **Verification Steps** - How to confirm the work is complete:
   - [ ] `make test` passes
   - [ ] `make lint` passes
   - [ ] Manual verification steps (if applicable)

5. **Acceptance Criteria Mapping** - Show how each CWO criterion will be met

### Step 2.4: Exit Plan Mode

When your plan is ready, use `ExitPlanMode` to present it for El Presidente's approval.

---

## PHASE 3: After Approval (CRITICAL)

When El Presidente approves your plan, you MUST immediately capture it to the ORC ledger before beginning any implementation.

### Step 3.1: Capture Plan to Ledger

```bash
./orc plan create "Your Plan Title" \
  --cycle-id CYC-XXX \
  --commission COMM-XXX \
  --shipment SHIP-XXX \
  --content "$(cat ~/.claude/plans/<plan-file>.md)"
```

Use the Cycle ID, Commission ID, and Shipment ID from Phase 1.

### Step 3.2: Approve the Plan

```bash
./orc plan approve PLAN-XXX
```

Use the PLAN ID returned from the create command.

### Step 3.3: Confirm Success

```bash
./orc plan show PLAN-XXX
```

Verify the plan is captured with status `approved`.

---

## PHASE 4: Begin Implementation

Only after the plan is captured and approved in the ORC ledger may you begin implementation.

Follow your plan. Follow AGENTS.md. Run verification at the end.

**After implementation is complete:**
Run `/orc-deliver` to close out the cycle.

---

## Quick Reference

```
PREREQUISITE: CWO exists and approved (from /orc-cycle)
    │
    ▼
/orc-plan
    │
    ▼
PHASE 1: Context (orc status, AGENTS.md, WO, CWO)
    │
    ▼
PHASE 2: EnterPlanMode → Design → ExitPlanMode
    │
    ▼
El Presidente Approves
    │
    ▼
PHASE 3: orc plan create → orc plan approve
    │
    ▼
PHASE 4: Implementation
    │
    ▼
CREC created → El Presidente verifies
    │
    ▼
HANDOFF: /orc-cycle (reflects CREC, plucks next cycle)
```
