---
name: imp-rec
description: Verify work, create receipt, complete task, chain to next task.
next: imp-plan-create
---

# IMP Receipt

Verify completed work, create a receipt, complete the task, and chain to the next task.

## Build System Detection

Infer test/lint commands from build system:

| Build System | Test Command | Lint Command |
|--------------|--------------|--------------|
| Makefile | `make test` | `make lint` |
| package.json | `npm test` | `npm run lint` |
| Gemfile | `bundle exec rspec` | `bundle exec rubocop` |

If no build system found:
```
Warning: No build system detected. Skipping automated verification.
Proceed with manual verification or skip if not applicable.
```

## Flow

1. **Verify work against plan**
   - Review git diff against plan (`git diff`)
   - Run tests (detected command or skip with warning)
   - Run lint (detected command or skip with warning)
   - Check any manual verification steps from plan

2. **Commit changes**
   ```
   /commit
   ```
   This commits verified work before creating the receipt. If there are no changes to commit, proceed to the next step.

3. **Create receipt**
   ```bash
   orc rec create "<outcome>" --task TASK-xxx
   ```

4. **Submit and verify receipt**
   ```bash
   orc rec submit REC-xxx
   orc rec verify REC-xxx
   ```

5. **Complete task**
   ```bash
   orc task complete TASK-xxx
   ```

6. **Check for next task**
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
   Output: "All tasks complete! Shipment ready for deploy. Run /ship-deploy." or "Waiting on blocked tasks."

## Verification Failure

If tests or lint fail:
- Do NOT complete the task
- Fix the issues
- Re-run verification
- Then proceed with receipt

## Notes

- Never skip verification steps (unless no build system detected)
- Receipt documents what was actually done
- Chaining to next task maintains propulsion
