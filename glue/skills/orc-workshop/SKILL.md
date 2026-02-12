---
name: orc-workshop
description: Guide through creating a new workshop with workbenches. Use when user says /orc-workshop or wants to create a new workshop for a project.
---

# Workshop Creation Skill

Interactive skill that walks users through creating a complete workshop with workbenches.

## Usage

```
/orc-workshop                           # Interactive - prompts for name and template
/orc-workshop "Workshop purpose"        # Prompts for template
/orc-workshop standard "Workshop name"  # Uses template directly
/orc-workshop muster "Feature work"     # Uses muster template
```

## Philosophy

This skill is an expert **user** of the ORC CLI, not an expert builder. If commands fail or syntax is unclear, consult `--help`:

```bash
orc workshop create --help
orc workbench create --help
orc tmux apply --help
```

## Templates

Templates are stored in `~/.orc/workshop-templates.json`. Each template defines a list of repos to create workbenches for.

**Default templates:**
- `standard` - intercom, infrastructure
- `muster` - intercom, muster, muster-deployer
- `ami` - intercom, ami, infrastructure
- `provisioner` - intercom, intercom-provisioner, infrastructure
- `events` - intercom, infrastructure, event-management-system

Use `/orc-workshop-templates` to manage templates.

## Flow

### Step 1: Parse Arguments and Load Templates

**Load templates:**
```bash
cat ~/.orc/workshop-templates.json
```

If file doesn't exist, create it with defaults:
```json
{
  "templates": {
    "standard": ["intercom", "infrastructure"],
    "muster": ["intercom", "muster", "muster-deployer"],
    "ami": ["intercom", "ami", "infrastructure"],
    "provisioner": ["intercom", "intercom-provisioner", "infrastructure"],
    "events": ["intercom", "infrastructure", "event-management-system"]
  }
}
```

**Parse arguments:**
- If first word matches a template name → use that template, rest is workshop name
- If first word doesn't match a template → entire string is workshop name, will prompt for template
- If no arguments → prompt for both

### Step 2: Workshop Purpose

If workshop name not determined from arguments, ask:

> "What's this workshop for? (e.g., 'orc development', 'DLQ admin tool', 'auth refactor')"

The answer becomes the workshop name.

### Step 3: Template Selection

If template not determined from arguments, show template menu:

> "Use a template? (default: standard)"
>
> 1. **standard** - intercom, infrastructure
> 2. **muster** - intercom, muster, muster-deployer
> 3. **ami** - intercom, ami, infrastructure
> 4. **provisioner** - intercom, intercom-provisioner, infrastructure
> 5. **events** - intercom, infrastructure, event-management-system
> 6. **[manual]** - select repos individually

If user presses enter or selects 1, use "standard" template.
If user selects "[manual]", skip to Step 5 (manual repo selection).

### Step 4: Create Workshop

```bash
orc workshop create --name "<purpose>"
```

Capture the created `WORK-xxx` ID from output.

### Step 5: Resolve Template Repos

If using a template:

1. Get repo list: `orc repo list`
2. For each repo name in template, find matching REPO-xxx ID
3. If repo not found, help user register it:

> "Repo 'muster' not found. Let me help you register it."
> "What's the local path to muster? (e.g., ~/src/muster)"

```bash
orc repo create "muster" --path "<path>"
```

Then retry lookup.

**If manual selection (no template):**

```bash
orc repo list
```

Display the list and ask which repos they need.

### Step 6: Create Workbenches

For each repo in template (or manual selection):

```bash
orc workbench create --workshop WORK-xxx --repo-id REPO-yyy
```

The name is auto-generated as `{repo}-{number}` (e.g., `intercom-015`).

### Step 7: Start TMux Session

```bash
orc tmux apply WORK-xxx --yes
```

This creates/reconciles the tmux session for the workshop, including windows for each workbench with the standard pane layout.

Ask: "Does this look right? Want to add/remove any workbenches?"

**Iterate as needed:**
- Add more workbenches → repeat Step 6, then re-run `orc tmux apply WORK-xxx --yes`
- Remove a workbench → `orc workbench archive BENCH-xxx`, re-run `orc tmux apply WORK-xxx --yes`
- Satisfied → proceed to commission linking

### Step 8: Commission Linking

```bash
orc commission list
```

Show active commissions and ask:

> "Link this workshop to a commission? Pick one, or create a new commission."

**If linking to existing:**
```bash
orc workshop set-commission WORK-xxx COMM-yyy
```
(Note: Must run from coordinator directory. Link after workbenches are created - see below.)

**If creating new:**
```bash
orc commission create "<title>" --description "<description>"
```

**If skipping:**
Warn: "Workshop will be created without a commission. IMPs/goblins won't have focus context. You can link later with `orc workshop set-commission`."

**After workbenches created, if commission was selected:**
```bash
cd ~/.orc/ws/WORK-xxx-*/
orc workshop set-commission COMM-yyy
```

### Step 9: Confirm Ready

Output summary:

```
Workshop created:
  WORK-xxx: <purpose>
  Template: <template-name> (or "manual")
  Workbenches:
    - BENCH-xxx: <repo>-xxx (~/wb/<name>)
    - BENCH-yyy: <repo>-yyy (~/wb/<name>)
  Commission: COMM-xxx (or "not linked")

To start working:
  orc tmux connect WORK-xxx
```

## Example Session (with template)

```
User: /orc-workshop standard "DLQ admin tool"

Agent: Creating workshop "DLQ admin tool" with template "standard"...
       [runs orc workshop create --name "DLQ admin tool"]
       ✓ Created WORK-005

       Template "standard" includes: intercom, infrastructure
       [looks up repo IDs]
       [runs orc workbench create for each]
       ✓ Created BENCH-016: intercom-016
       ✓ Created BENCH-017: infrastructure-017

       [runs orc tmux apply WORK-005 --yes]
       ✓ TMux session created

       Link to a commission?
       [continues with commission linking...]
```

## Example Session (interactive)

```
User: /orc-workshop

Agent: What's this workshop for?

User: DLQ admin tool

Agent: Use a template? (default: standard)
       1. standard - intercom, infrastructure
       2. muster - intercom, muster, muster-deployer
       3. ami - intercom, ami, infrastructure
       4. provisioner - intercom, intercom-provisioner, infrastructure
       5. events - intercom, infrastructure, event-management-system
       6. [manual] - select repos individually

User: 1

Agent: Creating workshop "DLQ admin tool" with template "standard"...
       [continues as above]
```

## Example Session (missing repo)

```
User: /orc-workshop muster "Muster feature"

Agent: Creating workshop "Muster feature" with template "muster"...
       ✓ Created WORK-006

       Template "muster" includes: intercom, muster, muster-deployer
       [looks up repo IDs]

       Repo 'muster-deployer' not found. Let me help you register it.
       What's the local path to muster-deployer?

User: ~/src/muster-deployer

Agent: [runs orc repo create "muster-deployer" --path "~/src/muster-deployer"]
       ✓ Registered REPO-008: muster-deployer

       [continues with workbench creation...]
```

## Error Handling

| Error | Remediation |
|-------|-------------|
| Templates file missing | Create with defaults |
| Repo not found in template | Help user register it |
| Path already exists | Suggest removing directory or different name |
| TMux apply fails | Show error, offer specific fix based on message |
| Commission not found | Offer to create new commission |

## CLI Reference

Always consult `--help` for current syntax:

| Command | Purpose |
|---------|---------|
| `orc workshop create --help` | Create workshop |
| `orc workbench create --help` | Create workbench |
| `orc repo list` | List available repos |
| `orc repo create --help` | Register new repo |
| `orc tmux apply WORK-xxx --yes` | Create/reconcile tmux session |
| `orc commission list` | List commissions |
| `orc workshop set-commission --help` | Link workshop to commission |
| `orc tmux connect WORK-xxx` | Attach to workshop tmux session |
