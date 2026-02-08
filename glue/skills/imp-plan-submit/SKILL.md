---
name: imp-plan-submit
description: Submit and approve the current plan.
next: imp-implement
---

# IMP Plan Submit

Submit and approve the current plan, then proceed to implementation.

## Flow

1. **Get draft plan for current task**
   ```bash
   orc plan list --task TASK-xxx --status draft
   ```

2. **Submit and approve plan**
   ```bash
   orc plan submit PLAN-xxx
   orc plan approve PLAN-xxx
   ```

3. **Output**
   "Plan PLAN-xxx approved. Run /imp-implement to see the plan and begin coding."

## Notes

- Plans are approved immediately - no review step
- Task specification happens upstream via /ship-plan
- If truly stuck during implementation, use /imp-escalate
