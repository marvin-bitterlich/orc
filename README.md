# ORC - The Forest Factory

![Forest Factory](assets/orc.png)

Deep in the forest stands a factory. Goblins coordinate from their workbench panes while IMPs -- disposable worker agents -- hammer out code. Shipments move through the system: bundles of tasks ready for delivery.

ORC is a CLI for structured AI-assisted development. It tracks commissions, organizes work into containers, preserves context across sessions, and provisions isolated workspaces. The forest runs on SQLite and git worktrees.

## Why ORC

- **No repo pollution** - Workbenches are git worktrees; state lives in external SQLite database
- **Ephemeral infrastructure** - tmux sessions spin up and tear down cleanly
- **Persistent ledger** - All work tracked in SQLite: commissions, shipments, tasks, plans
- **Semantic health tools** - docs-doctor validates documentation against code reality
- **Claude Teams integration** - ORC provides memory and policy; Teams provides execution

## Codebase Philosophy

- **Hex architecture with linting** - Strict layer boundaries enforced by go-arch-lint
- **Comprehensive testing** - Table-driven tests, repository tests, service tests
- **Git hooks as gates** - Pre-commit enforces lint and tests; no bypassing quality

## Getting Started

```bash
# Clone to the canonical location (required for orc doctor)
git clone <repo-url> ~/src/orc
cd ~/src/orc
make bootstrap
```

**Note:** ORC expects to live at `~/src/orc`. The `orc doctor` command validates this location, and some Make targets enforce it.

Then run `/orc-first-run` in Claude Code for an interactive walkthrough.

→ See [docs/getting-started.md](docs/getting-started.md) for detailed setup and first-run guide.

## Workflows

ORC follows a structured workflow with simple, manual lifecycles. Shipments progress through: draft -> ready -> in-progress -> closed.

-> See [docs/common-workflows.md](docs/common-workflows.md) for detailed workflow patterns.

## Glossary

| Term | Description |
|------|-------------|
| **Commission** | Grand undertaking that gives work context and purpose |
| **Shipment** | Bundle of tasks moving through the system (draft -> ready -> in-progress -> closed) |
| **Workbench** | Git worktree where agents do isolated work |
| **Goblin** | Coordinator agent -- human's long-running workbench pane |
| **IMP** | Disposable worker agent spawned by Claude Teams |
| **Task** | Deed to be done within a shipment (open -> in-progress -> closed) |

-> See [docs/schema.md](docs/schema.md) for complete terminology.

## The Cast

**Goblins** are coordinators -- the human's long-running workbench pane. They manage ORC tasks and provide memory and policy.

**IMPs** are disposable workers -- spawned by Claude Teams to execute tasks.

**Workbenches** are git worktrees -- isolated workspaces where changes happen safely.

-> See [docs/schema.md](docs/schema.md) for role details.

---

*The forest hums with industry. Shipments move through workbenches. Goblins coordinate. IMPs hammer at their tasks. The system remembers everything.*
