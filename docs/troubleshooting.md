# Troubleshooting

**Status**: Living document
**Last Updated**: 2026-02-08

Common issues and their solutions. This document grows with usage.

## Quick Help

Don't know which command to use? Run:

```
/orc-help
```

This skill provides:
- Categorized skill list (ship-*, imp-*, orc-*, goblin-*)
- Usage examples for common workflows
- Links to relevant documentation

## Common Issues

### Database

**Error: "database is locked"**

SQLite database is being accessed by multiple processes.

Solution:
1. Check for running ORC processes: `ps aux | grep orc`
2. Wait for other operations to complete
3. If stuck, restart: `orc doctor`

**Error: "no such table"**

Database schema is out of sync.

Solution:
```bash
make schema-apply   # Apply latest schema
orc doctor          # Verify database health
```

### Git / Worktrees

**Error: "worktree already exists"**

Attempting to create a worktree for a branch that already has one.

Solution:
```bash
git worktree list                    # Find existing worktrees
git worktree remove <path>           # Remove if unused
orc infra apply WORK-xxx             # Re-apply infrastructure
```

**Error: "not a git repository"**

Running ORC commands outside a git repository.

Solution:
- Navigate to a git repository
- Or create a workbench with `orc workbench create`

### Skills

**Skill not found**

The skill isn't deployed to Claude Code.

Solution:
```bash
make deploy-glue    # Redeploy all skills
```

**Skill not working as expected**

The skill may have been updated but not redeployed.

Solution:
```bash
make deploy-glue    # Hot reload all skills
```

### Hooks

**Stop hook blocking unexpectedly**

The Stop hook blocks when a shipment is in `auto_implementing` mode with incomplete tasks.

Solution:
- Complete the current task and let IMP chain to next
- Or switch to manual mode: `/imp-auto off`

**Hooks not firing**

Hooks may not be configured in settings.json.

Solution:
```bash
make deploy-glue    # Redeploys hooks.json configuration
cat ~/.claude/settings.json | jq '.hooks'   # Verify hooks present
```

### Infrastructure

**TMux session not found**

Workshop tmux session doesn't exist.

Solution:
```bash
orc infra apply WORK-xxx    # Creates tmux session
orc tmux connect WORK-xxx   # Connect to session
```

**Workbench directory missing**

Worktree was deleted but database record remains.

Solution:
```bash
orc infra apply WORK-xxx    # Recreates missing worktrees
```

## FAQ

### How do I find the right skill?

Run `/orc-help` for a categorized overview. Skills follow naming conventions:
- `/ship-*` - Shipment lifecycle
- `/imp-*` - Implementation workflow
- `/orc-*` - Utilities

### How do I reset my session?

If Claude Code is in a confused state:
1. Exit and restart Claude Code
2. Run `orc prime` for fresh context

### How do I check system health?

```bash
orc doctor    # Environment and database health
orc status    # Current context and focus
orc summary   # Full commission tree
```

### Where are ORC files stored?

| Location | Purpose |
|----------|---------|
| `~/.orc/orc.db` | SQLite database |
| `~/.orc/config.json` | Global configuration |
| `~/.claude/skills/` | Deployed skills |
| `~/.claude/hooks/` | Deployed hooks |
| `.orc/config.json` | Workbench identity |

### How do I update ORC?

```bash
git pull               # Get latest code
make install           # Rebuild and install
make deploy-glue       # Update skills and hooks
```

## Getting Help

If you encounter an issue not covered here:

1. Run `orc doctor` to check system health
2. Check `~/.claude/orc-debug.log` for recent activity
3. Review the [architecture docs](architecture.md) for system understanding
4. Create a note: `orc note create "Issue: description" --shipment SHIP-xxx`
