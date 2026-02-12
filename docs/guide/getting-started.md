# Getting Started with ORC

**Status**: Living document
**Last Updated**: 2026-02-12

Setup happens in three phases: **clone**, **install**, and **onboard**. Each phase builds on the previous one.

## Prerequisites

Before installing ORC, ensure you have:

| Requirement | Version | Check Command |
|-------------|---------|---------------|
| Homebrew | any | `brew --version` |
| git | 2.x+ | `git --version` |
| make | any | `make --version` |
| Claude Code | latest | `claude --version` |

**Note:** Go is installed automatically via Homebrew during the install phase.

## Phase 1: Clone

ORC expects to live at `~/src/orc`. This location is validated by `orc doctor` and used by the install phase to register the repository.

```bash
mkdir -p ~/src
git clone <repo-url> ~/src/orc
cd ~/src/orc
```

**What you have after this phase:** Source code on disk, nothing installed.

## Phase 2: Install (`make bootstrap`)

```bash
make bootstrap
```

This sets up the machine — tools, binary, database, and configuration. Specifically:

- Installs dependencies via `brew bundle` (including Go)
- Builds and installs the `orc` binary to `$GOPATH/bin`
- Installs git hooks (pre-commit quality gates)
- Deploys Claude Code skills and hooks
- Initializes the SQLite database at `~/.orc/orc.db`
- Creates directories (`~/.orc/ws/`, `~/wb/`)
- Configures PATH (adds `$GOPATH/bin` to `~/.zprofile` and `~/.zshrc`)
- Creates FACT-001 (default factory) and REPO-001 (ORC repository)
- Runs `orc doctor` to verify the installation

For development dependencies (VM testing, schema migrations):
```bash
make bootstrap-dev
```

**What you have after this phase:** A working `orc` binary, database, and infrastructure — but no commissions, workshops, or shipments yet. The install phase prints: *"Next step: Run 'orc bootstrap' to start the first-run experience"*.

## Phase 3: Onboard (`orc bootstrap`)

```bash
orc bootstrap
```

This launches an interactive Claude Code session with the `/orc-first-run` skill. Where `make bootstrap` set up the machine, this phase sets up your workflow — creating the entities you need to start working:

- Creates a "Getting Started" commission (your first project)
- Creates a workshop with a workbench (your tmux workspace and git worktree)
- Creates a "First Steps" shipment and focuses on it
- Explains ORC concepts as it goes
- Guides you through optional repo and template configuration
- Ends with a quick-reference card of essential commands

The skill is adaptive — it checks what already exists before creating anything, so it's safe to run multiple times.

### Manual Alternative

If you prefer to set up your workflow manually instead of running `orc bootstrap` (note: Phase 2 already created FACT-001 and REPO-001):

```bash
# Create a commission
orc commission create "My First Project"

# Create a workshop
orc workshop create "Development" --factory FACT-001

# Create a workbench (git worktree)
orc workbench create my-workbench --workshop WORK-001 --repo-id REPO-001

# Apply infrastructure (creates tmux session and worktree)
orc infra apply WORK-001

# Connect to the workshop
orc tmux connect WORK-001
```

## Verification

After completing all three phases, verify everything is working:

```bash
orc doctor     # Check environment health
orc status     # View your current context
orc summary    # See the commission tree
```

## Next Steps

Now that ORC is installed:

1. **Learn the workflows** → [docs/guide/common-workflows.md](common-workflows.md)
2. **Understand the architecture** → [docs/reference/architecture.md](../reference/architecture.md)
3. **Review terminology** → [docs/guide/glossary.md](glossary.md)
