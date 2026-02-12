# CLAUDE.md - ORC Development Guide

Essential rules and pointers for agents working on the ORC codebase.

## Essential Commands

```bash
make dev        # Build local ./orc (always use ./orc for development)
make test       # Run all tests
make lint       # golangci-lint + go-arch-lint (architecture boundaries)
make init       # Install git hooks after cloning
```

```bash
orc prime       # Context injection at session start
orc status      # Current commission and shipment status
orc summary     # Hierarchical view of shipments and tomes with pinned items
orc doctor      # Validate ORC environment and glue deployment
```

Skills: `/release`, `/ship-deploy` (see [docs/dev/deployment.md](docs/dev/deployment.md)), `/docs-doctor`

## Pre-Commit Checks (Enforced by Hook)

All commits must pass `make lint` (check-test-presence, check-coverage, schema-check, golangci-lint, go-arch-lint). Emergency bypass: `git commit --no-verify` (audited). Before merging to master, run `/docs-doctor`.

## Architecture Boundaries

Hexagonal (ports & adapters) with strict layers. **The linter config (`.go-arch-lint.yml`) is the source of truth.**

| Layer | Location | Rule |
|-------|----------|------|
| Core | `internal/core/` | Pure domain logic. No non-core imports. |
| Ports | `internal/ports/` | Interfaces only. Stdlib only. |
| App | `internal/app/` | Orchestration via core + ports. No direct I/O. |
| Adapters | `internal/adapters/` | Translation and I/O only. No business logic. |
| CLI | `internal/cli/` | Thin: parse args, call services, render output. |
| Wire | `internal/wire/` | Dependency injection only. |

Run `make lint` to verify. See [docs/architecture.md](docs/architecture.md) for full details.

## Common Mistakes

- Writing business logic in adapters (keep them as pure translation)
- Importing adapters/infra from core (core has no non-core imports)
- Calling tmux directly from app (use a port; adapter executes tmux)
- Claiming checks passed without running them (run and report explicitly)
- Running Atlas commands manually (use Makefile targets: `make schema-diff`, `make schema-apply`)

## Shipment & Task Lifecycles

Shipment: `draft` -> `ready` -> `in-progress` -> `closed` (all transitions manual)
Task: `open` -> `in-progress` -> `blocked` -> `closed` (all transitions manual)

See [docs/common-workflows.md](docs/common-workflows.md).

## Detailed Documentation

### For Agents (developing ORC)
- [docs/dev/checklists.md](docs/dev/checklists.md) -- Add field, add entity, add CLI command, add state
- [docs/dev/database.md](docs/dev/database.md) -- Atlas workflow, two-database model, schema changes
- [docs/dev/testing.md](docs/dev/testing.md) -- Table-driven tests, test pyramid, verification discipline
- [docs/dev/integration-tests.md](docs/dev/integration-tests.md) -- Integration test skills: coverage map and gaps
- [docs/dev/config.md](docs/dev/config.md) -- Config format, infrastructure plan/apply
- [docs/dev/deployment.md](docs/dev/deployment.md) -- Deployment workflow and checks
- [docs/dev/release.md](docs/dev/release.md) -- Release process
- [docs/dev/glue.md](docs/dev/glue.md) -- Skills and hooks system

### General
- [docs/getting-started.md](docs/getting-started.md) -- Setup and first-run
- [docs/common-workflows.md](docs/common-workflows.md) -- Shipment and task workflows
- [docs/schema.md](docs/schema.md) -- Schema, glossary, and terminology
- [docs/tmux.md](docs/tmux.md) -- TMux session management
- [docs/troubleshooting.md](docs/troubleshooting.md) -- Common issues
- [docs/architecture.md](docs/architecture.md) -- C2/C3 codebase structure
