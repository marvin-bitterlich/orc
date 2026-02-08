---
name: imp-plan-create
description: Research codebase and create implementation plan for current task. Uses Explore agent.
---

# IMP Plan Create

Research the codebase and create an implementation plan for the current in_progress task.

## Documentation Discovery

Look for development checklists and guidelines in order:
1. `CLAUDE.md` - Development rules and checklists
2. `docs/` - Additional documentation (architecture, guides)

If checklists found, reference them in the plan.

## Build System Detection

Infer verification commands from build system:

| Build System | Test Command | Lint Command |
|--------------|--------------|--------------|
| Makefile | `make test` | `make lint` |
| package.json | `npm test` | `npm run lint` |
| Gemfile | `bundle exec rspec` | `bundle exec rubocop` |

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

4. **Check for development checklists**
   Look for CLAUDE.md or docs/ with relevant checklists.

   If found, identify the relevant change type:
   | Change Type | Look For |
   |-------------|----------|
   | New entity | "Add Entity" checklist |
   | Add column | "Add Column" or "Add Field" checklist |
   | CLI command | "Add CLI Command" checklist |
   | State/transition | "Add State" checklist |

   If checklists found, your plan SHOULD follow them.
   If no checklists found, proceed with standard approach.

5. **Design implementation approach**
   Based on research, design a concrete approach:
   - What files to modify/create
   - What changes to make
   - How to verify

6. **Create plan record**
   ```bash
   orc plan create --task TASK-xxx "Brief plan title"
   ```

7. **Write plan content**
   Use detected verification commands:
   ```bash
   orc plan update PLAN-xxx --content "$(cat <<'EOF'
   ## Summary
   [1-2 sentence overview]

   ## Changes
   - file1: [description]
   - file2: [description]

   ## Verification
   - [ ] Run tests: `<detected-test-command>`
   - [ ] Run lint: `<detected-lint-command>`
   - [ ] Manual verification: [if any]
   EOF
   )"
   ```

   If no build system detected:
   ```
   ## Verification
   - [ ] Manual verification: [describe checks]
   ```

8. **Output**
   "Plan PLAN-xxx created for TASK-xxx. Run /imp-plan-submit when ready."

## Guidelines

- Keep plans focused and scoped to the task
- Be concrete about file paths and changes
- Include specific verification steps
- Don't over-engineer or add scope creep
- **Follow discovered checklists** when available
