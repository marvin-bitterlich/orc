---
name: imp-rec
description: Verify work, create receipt, complete task, chain to next task.
---

# IMP Receipt

Verify completed work, create a receipt, complete the task, and chain to the next task.

## Flow

1. **Verify work against plan**
   - Review git diff against plan (`git diff`)
   - Run tests (`make test`)
   - Run lint (`make lint`)
   - Check any manual verification steps from plan

2. **Commit changes**
   ```
   /commit
   ```
   This commits verified work before creating the receipt. If there are no changes to commit, proceed to the next step.

3. **Create receipt**
   ```bash
   orc rec create --task TASK-xxx
   ```

4. **Add receipt content**
   ```bash
   orc rec update REC-xxx --content "$(cat <<'EOF'
   ## Changes Made
   - file1.go: [what was done]
   - file2.go: [what was done]

   ## Verification
   - [x] Tests pass
   - [x] Lint clean
   - [x] [other checks]

   ## Notes
   [any relevant notes]
   EOF
   )"
   ```

5. **Submit receipt**
   ```bash
   orc rec submit REC-xxx
   ```

6. **Verify and complete**
   If all verification passed:
   ```bash
   orc rec verify REC-xxx
   orc task complete TASK-xxx
   ```

7. **Check for next task**
   ```bash
   orc task list --shipment SHIP-xxx --status ready
   ```

   **If ready tasks found:**
   ```bash
   orc task claim TASK-yyy
   ```
   Output: "Task TASK-xxx completed. Claimed TASK-yyy. Run /imp-plan-create."

   **If no ready tasks:**
   Check if shipment is complete:
   ```bash
   orc shipment show SHIP-xxx
   ```
   Output: "All tasks complete! Shipment ready for deploy. Next: Run /ship-deploy to merge to master." or "Waiting on blocked tasks."

## Verification Failure

If tests or lint fail:
- Do NOT complete the task
- Fix the issues
- Re-run verification
- Then proceed with receipt

## Notes

- Never skip verification steps
- Receipt documents what was actually done
- Chaining to next task maintains propulsion
