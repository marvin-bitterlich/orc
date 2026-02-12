# Common Workflows

**Status**: Living document
**Last Updated**: 2026-02-11

This guide covers the standard patterns for working with ORC.

## Shipment Lifecycle

Shipments are the primary unit of work in ORC. They progress through a simple 4-status lifecycle.

### State Descriptions

| State | Description |
|-------|-------------|
| `draft` | Shipment created but not yet scoped |
| `ready` | Scoped and ready for implementation |
| `in-progress` | Active implementation |
| `closed` | Terminal state |

All transitions are manual -- the Goblin (coordinator) decides when to advance.

### Task Lifecycle

| State | Description |
|-------|-------------|
| `open` | Task created, available for work |
| `in-progress` | Actively being worked on |
| `closed` | Terminal state |

`blocked` is a lateral flag (not a status) that can be set on any non-closed task.

## Creating Work

### Starting a New Shipment

```
/ship-new "Title of the work"
```

Creates a shipment in `draft` status. Use for any piece of work you want to track.

### Quick Idea Capture

```
/orc-ideate
```

Rapid idea capture for brainstorming. Creates a focused shipment for quick exploration.

### Knowledge Synthesis

```
/ship-synthesize
```

When a shipment has accumulated exploration notes, use this to compact them into a summary note. Transforms messy exploration into structured knowledge.

### Planning Tasks

```
/ship-plan
```

C2/C3 engineering review that pressure-tests synthesized knowledge and creates tasks. Use when ready to convert exploration into actionable implementation.

## Workshop Management

### Setting the Active Commission

The active commission scopes the Goblin's focus and operations. It is stored in `workshops.active_commission_id`.

```bash
orc workshop set-commission COMM-001   # Set active commission
orc workshop set-commission --clear    # Clear active commission
```

## Goblin Workflow

The Goblin (coordinator) is the human's long-running workbench pane. It manages ORC tasks and context:

1. **Create shipments** with `/ship-new`
2. **Advance lifecycle** manually (draft -> ready -> in-progress -> closed)
3. **Create and manage tasks** within shipments
4. **Coordinate with IMPs** via Claude Teams

## IMP Workflow

IMPs (workers) are disposable agents spawned by Claude Teams. They execute tasks:

1. **Receive task assignment** from Teams
2. **Read ORC context** for requirements
3. **Implement changes** in their workbench
4. **Report completion** back to Teams

## Deployment

### Deploy Shipment

```
/ship-deploy
```

Merges the workbench branch to main. Only available when all tasks are closed.

### Complete Shipment

```
/ship-complete
```

Marks the shipment as closed after verification passes.

## Next Steps

- [docs/dev/glue.md](dev/glue.md) - Skills and hooks system
- [docs/troubleshooting.md](troubleshooting.md) - Common issues
- [docs/architecture.md](architecture.md) - System design
