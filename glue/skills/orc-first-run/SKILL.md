---
name: orc-first-run
description: Interactive first-run walkthrough for new users. Guides through creating first commission, workshop, and shipment.
---

# ORC First Run

Adaptive onboarding for new ORC users. Checks what already exists, creates what's missing, explains concepts as it goes, and guides through environment configuration.

## Usage

```
/orc-first-run
```

Run this via `orc hello` for the complete first-run experience.

When launched with `orc hello --factory FACT-xxx`, the skill will create
the workshop in the specified factory instead of 'default'.

## Philosophy

This skill is **adaptive** - it checks existing state before creating anything:
- If entities exist, tour and explain them
- If entities are missing, create them with explanations
- Never fail or create duplicates

## Flow

### Step 1: Welcome

Display welcome message:

```
Welcome to ORC - The Forest Factory!

ORC helps you orchestrate software development work through a hierarchy:

  Commission  → A body of work (project, initiative)
    Shipment  → A unit of work (feature, fix, exploration)
      Task    → An atomic piece of work

  Workshop    → A tmux session where agents work
    Workbench → A git worktree where an IMP implements tasks

Let me check what's already set up...
```

### Step 2: Check Existing State

Run discovery commands:

```bash
orc commission list
orc workshop list
orc repo list
```

Determine what exists:
- **Commission**: Look for one named "Getting Started" or any active commission
- **Workshop**: Look for any active workshop
- **Repo**: Look for "orc" repo or check if current directory is registered
- **Workbench**: If workshop exists, check for workbenches

### Step 3: Handle ORC Repo

The ORC repo should be registered so workbenches can be created.

```bash
orc repo list | grep -i orc
```

**If ORC repo not found:**
```bash
orc repo create orc --path ~/src/orc --default-branch main
```

Explain:
```
Registered the ORC repository at ~/src/orc (the canonical installation location).
This lets ORC create workbenches (git worktrees) for development.
```

**If ORC repo exists:**
```
Found ORC repository already registered.
```

### Step 4: Handle Commission

```bash
orc commission list | grep -i "getting started"
```

**If "Getting Started" commission not found:**
```bash
orc commission create "Getting Started" --description "Orientation commission for learning ORC"
```

Capture COMM-xxx ID.

Create Goblin orientation note:
```bash
orc note create "Goblin Orientation" --commission COMM-xxx --type spec --content "This is an orientation commission for new ORC users. Be extra explanatory about concepts and workflows as they explore."
```

Explain:
```
Created commission: COMM-xxx "Getting Started"

A commission groups related work together. This one is for learning ORC.
I've added a note to help the Goblin (workshop coordinator) know this is
an orientation context.
```

**If commission exists:**
```
Found existing commission: COMM-xxx

I'll use this for your first workshop and shipment.
```

### Step 5: Handle Workshop

```bash
orc workshop list
```

**If no workshop linked to the commission:**

If the directive includes a factory (e.g., "Use factory FACT-xxx"), pass it to the command:
```bash
orc workshop create --factory FACT-xxx
```

Otherwise use the default factory:
```bash
orc workshop create
```

Capture WORK-xxx ID.

Link to commission:
```bash
orc workshop set-commission COMM-xxx
```

Explain:
```
Created workshop: WORK-xxx

A workshop is a tmux session where you and your IMP agents work.
It's now linked to the "Getting Started" commission.
```

**If workshop exists:**
```
Found existing workshop: WORK-xxx

I'll use this for your first workbench.
```

### Step 6: Handle Workbench

```bash
orc workbench list --workshop WORK-xxx
```

**If no workbench exists:**
```bash
# Get ORC repo ID
orc repo list | grep orc
orc workbench create --workshop WORK-xxx --repo-id REPO-xxx
```

Capture BENCH-xxx ID.

Start tmux session:
```bash
orc tmux apply WORK-xxx --yes
```

Explain:
```
Created workbench: BENCH-xxx

A workbench is a git worktree where an IMP (implementation agent) works.
The tmux session has been created - directories and tmux windows are ready.
```

**If workbench exists:**
```
Found existing workbench: BENCH-xxx

Your workshop already has a workbench set up.
```

### Step 7: Handle Shipment

```bash
orc shipment list --commission COMM-xxx
```

**If no shipment exists:**
```bash
orc shipment create "First Steps" --commission COMM-xxx --description "Your first shipment - experiment and learn!"
```

Capture SHIP-xxx ID.

Focus the workbench on this shipment:
```bash
orc focus SHIP-xxx
```

Create user orientation note:
```bash
orc note create "Welcome" --shipment SHIP-xxx --type idea --content "Welcome to ORC! Use 'orc summary' to see your workspace, or start exploring with '/ship-new' to create your first real shipment."
```

Explain:
```
Created shipment: SHIP-xxx "First Steps"

A shipment tracks work through stages:
  draft → ready → in-progress → closed

This shipment is now focused - any notes or tasks you create will attach here.
```

**If shipment exists:**
```
Found existing shipment: SHIP-xxx

You already have a shipment to work with.
```

### Step 8: Show Current State

```bash
orc summary
```

Explain:
```
Here's your current ORC setup. The summary command shows your work hierarchy -
you'll use this constantly to see what's happening.
```

### Step 9: TMux Navigation

Display tmux explanation:

```
TMux Navigation

Your workshop runs in a tmux session with multiple windows:
- Window 0: Coordinator (Goblin - your workbench pane)
- Window 1+: Workbenches (where IMP workers operate)

Key bindings:
- Ctrl+b n     → Next window
- Ctrl+b p     → Previous window
- Ctrl+b w     → Window list
- Right-click  → ORC context menu (if configured)

Connect to your workshop with:
  orc tmux connect WORK-xxx
```

### Step 10: Guide Repo Configuration

Ask if user wants to add more repos:

```
Would you like to add more repositories for creating workbenches?

Common locations are ~/src/ or ~/code/. I can help you register repos
so you can create workbenches for them later.
```

Use AskUserQuestion:
- Question: "Would you like to add repositories from ~/src?"
- Header: "Repos"
- Options:
  1. "Yes, show me how" - Guide through /orc-repo
  2. "Skip for now" - Continue to templates
  3. "I'll do it later" - Continue to templates

**If user wants to add repos:**
```
Let's add some repositories. Run:
  /orc-repo

This will guide you through registering a repository.
After adding repos, you can create workbenches for them in any workshop.
```

Wait for user to complete /orc-repo or indicate they're done.

### Step 11: Guide Template Configuration

Ask about workshop templates:

```
Workshop templates let you quickly create workshops with pre-configured
workbenches. For example, a "standard" template might create workbenches
for your main repo plus infrastructure.
```

Use AskUserQuestion:
- Question: "Would you like to set up workshop templates?"
- Header: "Templates"
- Options:
  1. "Yes, show me" - Guide through /orc-workshop-templates
  2. "Skip for now" - Continue to finish
  3. "I'll configure later" - Continue to finish

**If user wants to configure templates:**
```
Let's set up some templates. Run:
  /orc-workshop-templates

This will let you create, edit, or view workshop templates.
```

Wait for user to complete or indicate they're done.

### Step 12: Finish with Quick Reference

Display completion message and reference card:

```
You're all set! Here's a quick reference:

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Essential Commands:
  orc summary              See all your work
  orc status               Current context
  orc focus SHIP-xxx       Focus on a shipment
  orc prime                Restore context (start of session)

Shipment Workflow:
  /ship-new "Title"        Create new shipment
  /ship-plan               Plan tasks from notes
  /ship-deploy             Deploy to master

Workshop Management:
  /orc-workshop            Create a new workshop
  /orc-repo                Register a repository
  orc tmux connect WORK-xxx Connect to workshop

Getting Help:
  /orc-help                Show all ORC skills
  orc doctor               Check environment health

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

To start working:
  orc tmux connect WORK-xxx

Happy building!
```

Replace WORK-xxx with the actual workshop ID discovered/created earlier.

## Error Handling

- If commands fail, explain what went wrong and suggest manual steps
- If user declines all optional steps, that's fine - finish gracefully
- If entities already exist in unexpected states, explain and continue

## Notes

- Uses AskUserQuestion for interactive prompts
- Chains to /orc-repo and /orc-workshop-templates as needed
- Idempotent - safe to run multiple times
- Designed to be launched by `orc hello`
