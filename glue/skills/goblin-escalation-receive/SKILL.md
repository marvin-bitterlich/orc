---
name: goblin-escalation-receive
description: Handle incoming escalation. Check CLAUDE.md compliance first, resolve if clear-cut, wait if judgment needed.
---

# Goblin Escalation Receive

Handle escalation from IMP or reviewer.

## Flow

1. **Read full context**
   ```bash
   orc escalation show ESC-xxx
   orc plan show PLAN-xxx
   orc task show TASK-xxx
   ```

2. **Check CLAUDE.md compliance FIRST**
   Before evaluating the escalation reason:

   a. What type of change is this? (entity, column, CLI, etc.)
   b. Read CLAUDE.md checklist for that change type
   c. Does the plan follow all required steps?

   If plan violates CLAUDE.md:
   ```bash
   orc escalation resolve ESC-xxx --outcome rejected \
     --resolution "Plan does not follow CLAUDE.md. Missing: [specific items]. Revise plan to include all checklist items."
   ```
   STOP HERE. Do not evaluate escalation reason.

3. **Evaluate escalation reason**
   Only after confirming CLAUDE.md compliance:

   **Is the answer documented?**
   - Check CLAUDE.md for guidance
   - Check codebase patterns

   **If clear-cut (answer in docs):**
   ```bash
   orc escalation resolve ESC-xxx --outcome approved \
     --resolution "Per CLAUDE.md: [specific guidance]"
   ```

   **If judgment call (not in docs):**
   Output to El Presidente:
   "ESC-xxx requires your input. The plan follows CLAUDE.md, but the escalation reason involves [judgment area]. Awaiting direction."

## Autonomy Boundaries

| Situation | Action |
|-----------|--------|
| Plan violates CLAUDE.md | REJECT autonomously |
| Answer is in CLAUDE.md | RESOLVE autonomously |
| Architectural decision | WAIT for El Presidente |
| Ambiguous requirements | WAIT for El Presidente |
| Trade-off judgment | WAIT for El Presidente |

## Key Principle

**Reading CLAUDE.md is mandatory.** It is step 2, before evaluating anything else.
