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
- Deploys skills and hooks to Claude Code

For development dependencies (VM testing, schema migrations):
```bash
make bootstrap-dev
```

### 3. Verify Installation

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

The easiest way to get started is the interactive first-run skill:

1. Open Claude Code in any directory
2. Run the skill:
   ```
   /orc-first-run
   ```

The walkthrough guides you through:
- Creating your first commission (a project to work on)
- Setting up a workshop (collection of workbenches)
- Creating a workbench (isolated git worktree)
- Focusing on your first shipment

### Manual Setup

If you prefer manual setup:

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

1. **Learn the workflows** → [docs/common-workflows.md](common-workflows.md)
2. **Understand the architecture** → [docs/architecture.md](architecture.md)
3. **Review terminology** → [docs/glossary.md](glossary.md)
