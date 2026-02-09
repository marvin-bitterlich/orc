# ORC - The Forest Factory

![Forest Factory](assets/orc.png)

Deep in the forest stands a factory. The ORC oversees operations from the command center while IMPs work in scattered workbenches, hammering out code. Shipments move through the system - bundles of tasks ready for delivery.

ORC is a CLI for structured AI-assisted development. It tracks commissions, organizes work into containers, preserves context across sessions, and provisions isolated workspaces. The forest runs on SQLite and git worktrees.

## Why ORC

- **No repo pollution** - Workbenches are git worktrees; state lives in external SQLite database
- **Ephemeral infrastructure** - tmux sessions spin up and tear down cleanly
- **Persistent ledger** - All work tracked in SQLite: commissions, shipments, tasks, plans, receipts
- **Semantic health tools** - docs-doctor validates documentation against code reality
- **Auto + interactive modes** - Run autonomously through shipments or pause for human oversight

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

ORC follows a structured workflow: exploration → synthesis → planning → implementation → verification.

→ See [docs/common-workflows.md](docs/common-workflows.md) for detailed workflow patterns.

## Glossary

| Term | Description |
|------|-------------|
| **Commission** | Grand undertaking that gives work context and purpose |
| **Shipment** | Bundle of tasks moving through the system |
| **Workbench** | Git worktree where an IMP does isolated work |
| **IMP** | Implementation agent that writes code |
| **Task** | Deed to be done within a shipment |
| **Handoff** | Context passed between sessions |

→ See [docs/glossary.md](docs/glossary.md) for complete terminology.

## The Cast

**The ORC** is your Orchestrator - oversees the forest, coordinates commissions, maintains the big picture.

**IMPs** are Implementation agents - inhabit workbenches and do the actual coding.

**Workbenches** are git worktrees - isolated workspaces where changes happen safely.

→ See [docs/glossary.md](docs/glossary.md) for role details.

---

*The forest hums with industry. Shipments move through workbenches. IMPs hammer at their tasks. And the ORC watches over all.*
