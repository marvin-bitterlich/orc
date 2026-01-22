# ORC Cycle Command

Cycle management ceremony - scopes **what** to build by analyzing the Work Order and creating CWOs.

## Overview

This skill helps you:
- Validate a Work Order has a coherent cycle plan
- Reflect completed cycle work back into the WO
- Pluck the next cycle chunk and create a CWO

**Boundary:** This skill is about **scoping work**, not planning implementation. It creates the CWO that defines *what* to build. The `/orc-plan` skill then plans *how* to build it.

**Workflow position:**
```
/orc-cycle → CWO created → El Presidente approves → /orc-plan → implement → CREC → /orc-cycle
```

> **Note:** This skill describes the target workflow. Some CLI commands may not exist yet (e.g., `orc work_order update`, Cycle FSM automation). If a command doesn't exist or fails, **stop and ask El Presidente for help** - they may want to build it, use a workaround, or skip that step.

---

## PHASE 1: Understand Current State

### Step 1.1: Get Context

```bash
./orc status
./orc work_order show WO-XXX
./orc cycle list --shipment-id SHIP-XXX
```

Understand:
- What is the WO outcome and acceptance criteria?
- Are there existing cycles? What's their status?
- Is there a cycle currently in progress?

### Step 1.2: Analyze the Work Order

The WO acceptance criteria should implicitly or explicitly contain a **cycle plan** - a logical breakdown of work into implementable chunks.

If the WO is new/fresh, the criteria might be a flat list that needs to be mentally grouped into cycles.

If cycles have already been run, some criteria may have been addressed.

---

## PHASE 2: Handle Previous Cycle (if applicable)

If the last cycle is complete (status: `complete`, CREC verified):

### Step 2.1: Read the CREC

```bash
./orc crec list --cycle-id CYC-XXX
./orc crec show CREC-XXX
```

### Step 2.2: Reflect Back to WO

Update the WO to note what was accomplished:
- Which acceptance criteria were addressed?
- Any scope changes or learnings?
- Use `orc work_order update` (with El Presidente's --confirm)

> If `orc work_order update` doesn't exist yet, stop and ask El Presidente how to proceed.

Ask El Presidente: "The last cycle completed. Here's what CREC-XXX says was done: [summary]. Should I update WO-XXX to reflect this?"

---

## PHASE 3: Pluck Next Cycle

### Step 3.1: Identify Next Chunk

Look at the WO acceptance criteria. Identify the next logical chunk of work that:
- Is coherent (related items that make sense together)
- Is achievable in one cycle
- Has clear acceptance criteria
- Builds on completed work (if any)

### Step 3.2: Propose to El Presidente

Present your proposed cycle scope:

"Based on WO-XXX, I propose the next cycle focus on:

**Outcome:** [what this cycle will achieve]

**Acceptance Criteria:**
1. [criterion 1]
2. [criterion 2]
...

This addresses WO criteria: [which ones]

Does this scope look right, El Presidente?"

### Step 3.3: Create Cycle and CWO

Once El Presidente approves:

```bash
# Create the cycle
./orc cycle create --shipment-id SHIP-XXX

# Create the CWO with approved scope
./orc cwo create "Cycle outcome here" \
  --cycle-id CYC-XXX \
  --acceptance-criteria '["criterion 1", "criterion 2"]'
```

---

## PHASE 4: Sanity Check

### Step 4.1: Validate Coherence

Before finishing, verify:
- [ ] CWO outcome is clear and achievable
- [ ] CWO criteria are specific and testable
- [ ] CWO scope aligns with WO goals
- [ ] No orphaned WO criteria (everything has a home in some cycle)

### Step 4.2: Report Status

```bash
./orc cycle show CYC-XXX
./orc cwo show CWO-XXX
```

Confirm to El Presidente:
- New cycle created: CYC-XXX
- CWO created: CWO-XXX (status: draft)

**Next step:** When El Presidente approves the CWO, run `/orc-plan` to design the implementation.

---

## Quick Reference

```
/orc-cycle
    │
    ▼
PHASE 1: Get context (status, WO, existing cycles)
    │
    ▼
PHASE 2: If last cycle complete → reflect CREC back to WO
    │
    ▼
PHASE 3: Analyze WO → propose next chunk → El Presidente approves → create CWO
    │
    ▼
PHASE 4: Sanity check → report
    │
    ▼
HANDOFF: El Presidente approves CWO → /orc-plan
```

---

## Edge Cases

**CLI command doesn't exist:**
- Stop immediately
- Tell El Presidente which command is missing
- Ask how to proceed (build it, workaround, or skip)

**Fresh WO, no cycles yet:**
- Skip Phase 2
- In Phase 3, propose the first cycle chunk

**WO criteria are vague:**
- Ask El Presidente to clarify before proceeding
- Help refine the WO if needed

**Cycle in progress (not complete):**
- Do NOT pluck a new cycle
- Inform El Presidente: "CYC-XXX is still in progress (status: X). Complete it first or mark it blocked/closed."

**All WO criteria addressed:**
- Congratulate El Presidente
- Suggest closing the shipment or adding more scope
