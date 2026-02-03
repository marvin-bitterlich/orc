---
name: imp-plan-create
description: Research codebase and create implementation plan for current task. Uses Explore agent.
---

# IMP Plan Create

Research the codebase and create an implementation plan for the current in_progress task.

## Flow

1. **Get in_progress task**
   ```bash
   orc task list --status implement
   ```
   Identify the task currently being worked on.

2. **Read task description**
   ```bash
   orc task show TASK-xxx
   ```
   Understand what needs to be done.

3. **Research codebase**
   Use the Task tool with `subagent_type: "Explore"` to research:
   - Find relevant files and patterns
   - Understand existing architecture
   - Identify affected areas

   Example prompt for Explore agent:
   "Research the codebase to understand how to implement: [task description]. Find relevant files, patterns, and architecture."

4. **Read AGENTS.md for change type**
   Identify what type of change this is and read the relevant checklist:

   | Change Type | AGENTS.md Section |
   |-------------|-------------------|
   | New entity | "Add Entity with Persistence" |
   | Add column | "Add Column to Existing Entity" |
   | CLI command | "Add CLI Command" |
   | State/transition | "Add State/Transition" |

   Your plan MUST follow the documented checklist.

5. **Design implementation approach**
   Based on research, design a concrete approach:
   - What files to modify/create
   - What changes to make
   - How to verify (tests, lint)

6. **Create plan record**
   ```bash
   orc plan create --task TASK-xxx "Brief plan title"
   ```

7. **Write plan content**
   ```bash
   orc plan update PLAN-xxx --content "$(cat <<'EOF'
   ## Summary
   [1-2 sentence overview]

   ## Changes
   - file1.go: [description]
   - file2.go: [description]

   ## Verification
   - [ ] Run tests: `make test`
   - [ ] Run lint: `make lint`
   - [ ] Manual verification: [if any]
   EOF
   )"
   ```

8. **Output**
   "Plan PLAN-xxx created for TASK-xxx. Run /imp-plan-submit when ready."

## Guidelines

- Keep plans focused and scoped to the task
- Be concrete about file paths and changes
- Include specific verification steps
- Don't over-engineer or add scope creep
- **Follow AGENTS.md checklists** for the change type
