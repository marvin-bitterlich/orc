# Getting Started with ORC

**Status**: Living document
**Last Updated**: 2026-02-08

This guide walks you through installing ORC and running your first session.

## Prerequisites

Before installing ORC, ensure you have:

| Requirement | Version | Check Command |
|-------------|---------|---------------|
| Homebrew | any | `brew --version` |
| git | 2.x+ | `git --version` |
| make | any | `make --version` |
| Claude Code | latest | `claude --version` |

**Note:** Go is installed automatically via Homebrew during bootstrap.

## Installation

### 1. Clone to the Canonical Location

ORC expects to live at `~/src/orc`. This location is validated by `orc doctor` and enforced by several Make targets.

```bash
mkdir -p ~/src
git clone <repo-url> ~/src/orc
cd ~/src/orc
```

### 2. Bootstrap the Environment

```bash
make bootstrap
```

This command:
- Verifies Homebrew is installed
- Installs dependencies via `brew bundle` (including Go)
- Builds the ORC binary
- Installs git hooks (pre-commit quality gates)
- Initializes the SQLite database at `~/.orc/orc.db`
- Creates FACT-001 (default factory) and REPO-001 (ORC repository)
- Deploys skills and hooks to Claude Code

For development dependencies (VM testing, schema migrations):
```bash
make bootstrap-dev
```

### 3. First-Run Setup (Recommended)

Launch the interactive first-run experience:

```bash
orc bootstrap
```

This opens Claude Code with the `/orc-first-run` skill, which guides you through:
- Creating your first commission
- Setting up a workshop with workbenches
- Focusing on your first shipment
- Connecting to tmux

Skip to "Verification" below after completing the walkthrough.

### 4. Verify Installation

```bash
orc doctor
```

Expected output shows all checks passing:
- Directory structure (`~/.orc/`, `~/wb/`)
- ORC repo at `~/src/orc`
- Glue deployment (skills, hooks)
- Hook configuration
- Binary installation

## First Run

### Interactive Walkthrough

The easiest way to get started is with `orc bootstrap`:

```bash
orc bootstrap
```

This launches Claude Code with the `/orc-first-run` skill. The walkthrough guides you through:
- Creating your first commission (a project to work on)
- Setting up a workshop (collection of workbenches)
- Creating a workbench (isolated git worktree)
- Focusing on your first shipment

### Manual Setup

If you prefer manual setup (note: `make bootstrap` already created FACT-001 and REPO-001):

```bash
# Create a commission
orc commission create "My First Project"

# Create a workshop
orc workshop create "Development" --factory FACT-001

# Create a workbench (git worktree)
orc workbench create my-workbench --workshop WORK-001 --repo-id REPO-001

# Apply infrastructure
orc infra apply WORK-001

# Connect to the workshop
orc tmux connect WORK-001
```

## Verification

After setup, verify everything is working:

```bash
# Check environment health
orc doctor

# View your context
orc status

# See the commission tree
orc summary
```

## Next Steps

Now that ORC is installed:

1. **Learn the workflows** → [docs/guide/common-workflows.md](common-workflows.md)
2. **Understand the architecture** → [docs/reference/architecture.md](../reference/architecture.md)
3. **Review terminology** → [docs/guide/glossary.md](glossary.md)
