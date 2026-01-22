# ORC Deliver Command

Delivery ceremony - closes out the current cycle by reviewing acceptance criteria and creating a CREC.

## Overview

This skill guides you through the structured delivery ceremony that:
- Reviews CWO acceptance criteria with El Presidente
- Creates a CREC (Cycle Receipt) documenting delivered outcomes
- Completes the CWO and submits the CREC for review
- Gets El Presidente's verification to complete the cycle

**Boundary:** This skill is about **closing the cycle**, not planning or scoping. The CWO was created by `/orc-cycle`, the plan by `/orc-plan`. This skill closes the loop.

**Workflow position:**
```
/orc-cycle → CWO created → El Presidente approves → /orc-plan → implement → /orc-deliver → /orc-cycle
```

**Prerequisite:** Implementation must be complete. The cycle must be in `implementing` status and the CWO must be `active`.

---

## PHASE 1: Context Gathering (MANDATORY)

### Step 1.1: Get Current State

```bash
./orc status
```

Note the focused **Shipment ID** and current **Cycle ID**.

### Step 1.2: Get Active CWO

```bash
./orc cwo list --shipment-id SHIP-XXX
./orc cwo show CWO-XXX
```

**Guards (must pass before proceeding):**
- [ ] Cycle status is `implementing`
- [ ] CWO status is `active`
- [ ] CWO does not already have a CREC

If guards fail, stop and inform El Presidente:
- If Cycle not `implementing`: "Cycle must be implementing. Current status: X. Run `/orc-plan` first?"
- If CWO not `active`: "CWO must be active. Current status: X"
- If CWO has CREC: "CWO already has CREC-XXX. This cycle has already been delivered."

---

## PHASE 2: Acceptance Criteria Review

### Step 2.1: Present Each Criterion

For each acceptance criterion in the CWO, present it to El Presidente and ask:

"**Criterion:** [the criterion text]

Was this criterion met? What's the evidence?"

### Step 2.2: Collect Responses

For each criterion, record:
- Met / Not Met / Partially Met
- Evidence (what was done, where to verify)
- Any notes or scope changes

### Step 2.3: Handle Unmet Criteria

If any criteria were not met:

Ask El Presidente: "Criterion X was not met. Options:
1. **Mark cycle blocked** - pause delivery, address the gap
2. **Adjust scope** - document the deviation, proceed with partial delivery
3. **Continue anyway** - note it and move forward

Which approach, El Presidente?"

Document the decision in the CREC evidence.

---

## PHASE 3: Create CREC

### Step 3.1: Compose Delivered Outcome

Based on Phase 2 responses, compose a summary of what was actually delivered:

```
Delivered Outcome: [concise summary of what was achieved]

Criteria Status:
1. ✅ [criterion 1] - [evidence]
2. ✅ [criterion 2] - [evidence]
...
```

### Step 3.2: Create the CREC

```bash
./orc crec create "Delivered outcome summary" \
  --cwo-id CWO-XXX \
  --evidence "Criteria Status:
1. ✅ criterion 1 - evidence
2. ✅ criterion 2 - evidence
..."
```

Note the returned **CREC ID** (e.g., CREC-005).

---

## PHASE 4: Complete CWO & Submit CREC

### Step 4.1: Complete the CWO

```bash
./orc cwo complete CWO-XXX
```

This marks the CWO as `completed`.

### Step 4.2: Submit the CREC

```bash
./orc crec submit CREC-XXX
```

This transitions the Cycle to `review` status.

### Step 4.3: Confirm State

```bash
./orc crec show CREC-XXX
./orc cycle show CYC-XXX
```

Confirm to El Presidente:
- CREC-XXX: status `submitted`
- CYC-XXX: status `review`

"Cycle is now in review. CREC is ready for your verification, El Presidente."

---

## PHASE 5: Verification

### Step 5.1: Present Delivery Summary

Display a summary for El Presidente:

```
┌────────────────────────────────────────────────┐
│ DELIVERY SUMMARY                               │
├────────────────────────────────────────────────┤
│ Cycle: CYC-XXX                                 │
│ CWO: CWO-XXX (completed)                       │
│ CREC: CREC-XXX (submitted)                     │
│                                                │
│ Delivered: [outcome summary]                   │
│                                                │
│ Evidence: [key evidence points]                │
└────────────────────────────────────────────────┘
```

### Step 5.2: Ask for Verification

"El Presidente, are you ready to verify this CREC and complete the cycle?
- **Yes** - CREC will be verified, cycle marked complete
- **Not yet** - CREC remains submitted for later verification"

### Step 5.3: Execute Verification (if approved)

If El Presidente says yes:

```bash
./orc crec verify CREC-XXX
```

This transitions the Cycle to `complete` status.

```bash
./orc cycle show CYC-XXX
```

Confirm: "Cycle CYC-XXX is now complete."

If El Presidente says not yet:
"CREC-XXX remains in `submitted` status. Verify later with `./orc crec verify CREC-XXX`."

---

## PHASE 6: Handoff

### Step 6.1: Report Final State

```bash
./orc status
```

Report to El Presidente:
- Cycle: CYC-XXX - [status]
- CWO: CWO-XXX - [status]
- CREC: CREC-XXX - [status]

### Step 6.2: Next Steps

If cycle is complete:
"The cycle is complete. Run `/orc-cycle` to:
- Reflect this work back to the WO
- Pluck the next chunk of work"

If cycle is in review (not yet verified):
"When ready to verify, run `./orc crec verify CREC-XXX` to complete the cycle."

---

## Quick Reference

```
PREREQUISITE: Implementation complete, Cycle: implementing, CWO: active
    │
    ▼
/orc-deliver
    │
    ▼
PHASE 1: Context (orc status, orc cwo show)
    │
    ▼
PHASE 2: Review each acceptance criterion with El Presidente
    │
    ▼
PHASE 3: Create CREC with delivered outcome and evidence
    │
    ▼
PHASE 4: orc cwo complete → orc crec submit → Cycle: review
    │
    ▼
PHASE 5: El Presidente verifies → orc crec verify → Cycle: complete
    │
    ▼
PHASE 6: Report final state
    │
    ▼
HANDOFF: /orc-cycle (reflects CREC to WO, plucks next chunk)
```

---

## Edge Cases

| Situation | Handling |
|-----------|----------|
| Cycle not `implementing` | Error: "Cycle must be implementing. Current: X" |
| CWO not `active` | Error: "CWO must be active. Current: X" |
| CWO already has CREC | Error: "CWO already has CREC-XXX. Cycle already delivered." |
| Criteria not met | Ask: "Mark cycle blocked or adjust scope?" |
| Scope changed during impl | Document in CREC evidence, note for WO update |
| CLI command fails | Stop, inform El Presidente, ask how to proceed |
