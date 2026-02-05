---
name: orc-workshop
description: Guide through creating a new workshop with workbenches. Use when user says /orc-workshop or wants to create a new workshop for a project.
---

# Workshop Creation Skill

Interactive skill that walks users through creating a complete workshop with workbenches.

## Usage

```
/orc-workshop
/orc-workshop "Workshop purpose"
```

## Philosophy

This skill is an expert **user** of the ORC CLI, not an expert builder. If commands fail or syntax is unclear, consult `--help`:

```bash
orc workshop create --help
orc workbench create --help
orc infra plan --help
orc infra apply --help
```

## Flow

### Step 1: Workshop Purpose

If purpose not provided as argument, ask:

> "What's this workshop for? (e.g., 'orc development', 'DLQ admin tool', 'auth refactor')"

The answer becomes the workshop name.

### Step 2: Create Workshop

```bash
orc workshop create --name "<purpose>"
```

Capture the created `WORK-xxx` ID from output.

**Note:** Always uses the default factory (no need to specify).

### Step 3: Show Available Repos

```bash
orc repo list
```

Display the list to the user.

**If no repos exist:**
> "No repos configured yet. Let's create one first."

Guide through repo creation:
```bash
orc repo create --help   # Check current syntax
orc repo create "<name>" --local-path "<path>"
```

### Step 4: Select Repos for Workbenches

Ask user which repos they need workbenches for.

**If user needs a repo not in the list:**
> "That repo isn't registered yet. Let me add it."

```bash
orc repo create "<name>" --local-path "<path>"
```

Then continue with workbench creation.

### Step 5: Create Workbenches

For each selected repo:

```bash
orc workbench create --workshop WORK-xxx --repo-id REPO-yyy
```

The name is auto-generated as `{repo}-{number}` (e.g., `intercom-015`).

### Step 6: Preview Infrastructure

```bash
orc infra plan WORK-xxx
```

Show the plan output to user. This displays:
- Gatehouse that will be created
- Workbenches (git worktrees) that will be created
- TMux windows that will be created

Ask: "Does this look right? Want to add/remove any workbenches?"

**Iterate as needed:**
- Add more workbenches → repeat Step 5
- Remove a workbench → `orc workbench archive BENCH-xxx`, re-run plan
- Satisfied → proceed to commission linking

### Step 7: Commission Linking

```bash
orc commission list
```

Show active commissions and ask:

> "Link this workshop to a commission? Pick one, or create a new commission."

**If linking to existing:**
```bash
orc workshop set-commission WORK-xxx COMM-yyy
```
(Note: Must run from gatehouse directory, but gatehouse doesn't exist yet. Use `cd ~/.orc/ws/WORK-xxx-*` won't work. Need to link after apply - see Step 8.)

**If creating new:**
```bash
orc commission create "<title>" --description "<description>"
```

**If skipping:**
Warn: "⚠️ Workshop will be created without a commission. IMPs/goblins won't have focus context. You can link later with `orc workshop set-commission`."

### Step 8: Apply Infrastructure

```bash
orc infra apply WORK-xxx
```

**If issues arise** (e.g., path already exists):
- Show the error
- Offer specific remediation:
  - Path exists: "Remove the directory or choose a different name"
  - Permission denied: "Check directory permissions"
  - Git worktree error: "The repo may have uncommitted changes"

**After successful apply, if commission was selected:**
```bash
cd ~/.orc/ws/WORK-xxx-*/
orc workshop set-commission COMM-yyy
```

### Step 9: Confirm Ready

Output summary:

```
Workshop created:
  WORK-xxx: <purpose>
  Gatehouse: GATE-xxx
  Workbenches:
    - BENCH-xxx: <repo>-xxx (~/wb/<name>)
    - BENCH-yyy: <repo>-yyy (~/wb/<name>)
  Commission: COMM-xxx (or "not linked")

To start working:
  orc tmux connect WORK-xxx
```

## Example Session

```
User: /orc-workshop "DLQ admin tool"

Agent: Creating workshop "DLQ admin tool"...
       [runs orc workshop create --name "DLQ admin tool"]
       ✓ Created WORK-005

       Which repos do you need workbenches for?
       [runs orc repo list]

       Available repos:
         REPO-001: intercom (/Users/me/src/intercom)
         REPO-002: orc (/Users/me/src/orc)

User: intercom

Agent: [runs orc workbench create --workshop WORK-005 --repo-id REPO-001]
       ✓ Created BENCH-016: intercom-016

       Any other repos?

User: no

Agent: Here's the infrastructure plan:
       [runs orc infra plan WORK-005]

       CREATE Gatehouse: ~/.orc/ws/WORK-005-dlq-admin-tool/
       CREATE Workbench: ~/wb/intercom-016 (git worktree)
       CREATE TMux window: intercom-016

       Look good?

User: yes

Agent: [runs orc commission list]

       Active commissions:
         COMM-001: ORC 3.1 Implementation
         COMM-002: DLQ Cleanup Project

       Link to a commission? (IMPs/goblins will start with this focus)

User: COMM-002

Agent: [runs orc infra apply WORK-005]
       ✓ Infrastructure created

       [runs cd ~/.orc/ws/WORK-005-*/ && orc workshop set-commission COMM-002]
       ✓ Workshop linked to commission

       Workshop created:
         WORK-005: DLQ admin tool
         Gatehouse: GATE-005
         Workbenches:
           - BENCH-016: intercom-016 (~/wb/intercom-016)
         Commission: COMM-002

       To start working:
         orc tmux connect WORK-005
```

## Error Handling

| Error | Remediation |
|-------|-------------|
| No repos exist | Guide through `orc repo create` |
| Repo not found | Offer to register it inline |
| Path already exists | Suggest removing directory or different name |
| Infra apply fails | Show error, offer specific fix based on message |
| Commission not found | Offer to create new commission |

## CLI Reference

Always consult `--help` for current syntax:

| Command | Purpose |
|---------|---------|
| `orc workshop create --help` | Create workshop |
| `orc workbench create --help` | Create workbench |
| `orc repo list` | List available repos |
| `orc repo create --help` | Register new repo |
| `orc infra plan WORK-xxx` | Preview infrastructure |
| `orc infra apply WORK-xxx` | Create infrastructure |
| `orc commission list` | List commissions |
| `orc workshop set-commission --help` | Link workshop to commission |
| `orc tmux connect WORK-xxx` | Attach to workshop tmux session |
