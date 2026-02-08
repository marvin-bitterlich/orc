# Glue System

**Status**: Living document
**Last Updated**: 2026-02-08

The glue system bridges Claude Code and the ORC CLI, providing skills and hooks that enable structured AI-assisted workflows.

## Overview

"Glue" refers to the integration layer between:
- **Claude Code** - The AI assistant running in your terminal
- **ORC CLI** - The orchestration system managing work

Skills are instructions for Claude. Hooks are scripts that fire on lifecycle events. Together they create a seamless workflow where Claude understands ORC context and follows established patterns.

## Directory Structure

```
glue/
├── skills/                 # Claude Code skill definitions
│   ├── ship-new/
│   │   └── SKILL.md
│   ├── imp-start/
│   │   └── SKILL.md
│   └── [30+ more skills]
│
├── hooks/                  # Lifecycle scripts
│   └── orc-debug-log.sh
│
├── hooks.json              # Hook configuration
│
├── tmux/                   # TMux integration
│   └── orc-session-picker.sh
│
└── README.md
```

## Skills

Skills are procedural instructions that Claude follows when invoked via `/skill-name`.

### File Format

Each skill is a directory containing a single `SKILL.md`:

```
glue/skills/<skill-name>/SKILL.md
```

### Anatomy

```markdown
---
name: skill-name
description: Clear description of when/why to use this skill. Include trigger words.
---

# Skill Title

## Usage
/skill-name           (basic)
/skill-name --flag    (with options)

## Flow

1. **Get context**
   ```bash
   orc status
   ```

2. **Perform action**
   ```bash
   orc <command>
   ```

3. **Output result**
   "Action complete: RESULT-xxx"

## Error Handling
- Condition not met → "Run `orc command` first."
```

### Frontmatter Fields

| Field | Purpose |
|-------|---------|
| `name` | Identifier used as `/name` in Claude Code |
| `description` | When/why to use, trigger words for discovery |

### Content Sections

| Section | Purpose |
|---------|---------|
| Usage | Command format with examples |
| Flow | Step-by-step instructions |
| Error Handling | How to handle failures |
| Guidelines | Best practices |

### Creating New Skills

1. Create directory: `glue/skills/my-skill/`
2. Create `SKILL.md` with frontmatter and instructions
3. Deploy: `make deploy-glue`
4. Test: `/my-skill` in Claude Code

Skills hot-reload after deployment - no restart needed.

## Hooks

Hooks are shell scripts that execute at Claude Code lifecycle events.

### Lifecycle Events

| Event | When It Fires |
|-------|---------------|
| `PreToolUse` | Before each tool call |
| `Stop` | When session stops |
| `SessionStart` | When session begins |
| `SessionEnd` | When session ends |

### Configuration

Hooks are configured in `glue/hooks.json`:

```json
{
  "PreToolUse": [
    {
      "matcher": "*",
      "hooks": [
        {
          "type": "command",
          "command": "~/.claude/hooks/orc-debug-log.sh",
          "timeout": 5000
        }
      ]
    }
  ],
  "Stop": [
    {
      "hooks": [
        {
          "type": "command",
          "command": "orc hook Stop",
          "timeout": 10000
        }
      ]
    }
  ]
}
```

### Hook Script Format

Hooks receive JSON on stdin and can:
- Log activity
- Block actions (return JSON with `decision: "block"`)
- Pass through silently

Example (debug logger):
```bash
#!/bin/bash
INPUT=$(cat)
echo "$INPUT" | jq -c '{tool: .tool_name}' >> ~/.claude/orc-debug.log
```

### ORC Stop Hook

The `orc hook Stop` command is special - it blocks session stop when:
- Shipment is in `auto_implementing` mode
- Tasks remain incomplete

This enables autonomous IMP workflow where the agent continues until the shipment is complete.

## Deployment

Deploy all glue components:

```bash
make deploy-glue
```

This command:
1. Copies skills to `~/.claude/skills/`
2. Copies hooks to `~/.claude/hooks/`
3. Merges `hooks.json` into `~/.claude/settings.json`
4. Copies tmux scripts to `~/.orc/tmux/`

The deploy is idempotent - safe to run multiple times.

### Deployment Targets

| Component | Destination |
|-----------|-------------|
| Skills | `~/.claude/skills/<name>/` |
| Hooks | `~/.claude/hooks/` |
| Hook config | `~/.claude/settings.json` |
| TMux scripts | `~/.orc/tmux/` |

## Discovery

### Finding Skills

Use the help skill to discover available commands:

```
/orc-help
```

Lists skill categories:
- **ship-*** - Shipment workflow
- **imp-*** - IMP (implementation) workflow
- **orc-*** - Utilities and management
- **goblin-*** - Review and escalation

### Skill Categories

| Prefix | Purpose |
|--------|---------|
| `ship-` | Shipment lifecycle (new, synthesize, plan, deploy, verify, complete) |
| `imp-` | Implementation workflow (start, plan-create, plan-submit, rec, auto) |
| `orc-` | Utilities (help, ping, interview, commission, workshop, workbench) |
| `goblin-` | Review workflow (escalation-receive) |

## Next Steps

- [docs/common-workflows.md](common-workflows.md) - Using skills in practice
- [docs/troubleshooting.md](troubleshooting.md) - Common issues
