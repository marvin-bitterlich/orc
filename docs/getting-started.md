# Getting Started with ORC

**Status**: Living document
**Last Updated**: 2026-02-08

This guide walks you through installing ORC and running your first session.

## Prerequisites

Before installing ORC, ensure you have:

| Requirement | Version | Check Command |
|-------------|---------|---------------|
| Go | 1.21+ | `go version` |
| git | 2.x+ | `git --version` |
| make | any | `make --version` |
| Claude Code | latest | `claude --version` |

## Installation

### 1. Clone the Repository

```bash
git clone <repo-url>
cd orc
```

### 2. Bootstrap the Environment

```bash
make bootstrap
```

This command:
- Builds the ORC binary
- Installs git hooks (pre-commit quality gates)
- Initializes the SQLite database at `~/.orc/orc.db`
- Deploys skills and hooks to Claude Code

### 3. Verify Installation

```bash
orc doctor
```

Expected output shows all checks passing:
- Database connection
- Git configuration
- Claude Code integration
- Skills deployment

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
orc workbench create my-workbench --workshop WORK-001 --repo <repo-path>

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
3. **Review terminology** → [docs/glossary/](glossary/)
