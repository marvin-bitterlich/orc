---
name: imp-implement
description: Minimal wrapper for coding phase. Shows approved plan, reminds IMP what to implement. User-driven exit (IMP calls /imp-rec when done).
next: imp-rec
---

# IMP Implement

Show the approved plan and remind IMP what to implement. Minimal wrapper for the coding phase.

## Flow

1. **Get current in-progress task**
   ```bash
   orc task list --status implement
   ```

2. **Get the approved plan for this task**
   ```bash
   orc plan list --task TASK-xxx --status approved
   orc plan show PLAN-xxx
   ```

3. **Display plan content to IMP**
   Show the full plan content so IMP knows what to implement.

4. **Output**
   "Now implement according to this plan. When done, run /imp-rec to verify and complete."

## Notes

- This is a minimal wrapper - no actual work happens here
- Helps IMP stay focused and know what to implement
- User-driven exit: IMP calls /imp-rec when implementation is complete
