# Documentation Schema

Defines the structure and validation rules for ORC's documentation. Used by `/docs-doctor` to validate the docs directory.

---

## Directory Structure

```
docs/
├── dev/              # Agent-facing: developing ORC itself
├── guide/            # Human-facing: using ORC as a product
└── reference/        # Both audiences: source-of-truth schemas and diagrams
```

### docs/dev/ — Development Documentation

**Audience**: Agents (IMPs, Goblins) developing ORC code.

**Purpose**: How to build, deploy, and modify ORC internals.

| File | Required | Description |
|------|----------|-------------|
| `checklists.md` | yes | Add Field, Add Entity, Add CLI, Add State checklists |
| `config.md` | yes | Config format, actor model, infrastructure plan/apply |
| `database.md` | yes | Atlas workflow, two-database model, schema changes |
| `deployment.md` | yes | Deployment checklist and merge-to-master process |
| `testing.md` | yes | Table-driven tests, test pyramid, verification discipline |

### docs/guide/ — User Guide

**Audience**: Humans using ORC as a product.

**Purpose**: Getting started, workflows, troubleshooting.

| File | Required | Description |
|------|----------|-------------|
| `getting-started.md` | yes | First-run setup and orientation |
| `common-workflows.md` | yes | Day-to-day workflow recipes |
| `troubleshooting.md` | yes | Common issues and fixes |
| `glossary.md` | yes | Term definitions (no mermaid diagrams) |
| `tmux.md` | no | TMux session management guide |

### docs/reference/ — Reference Documentation

**Audience**: Both agents and humans.

**Purpose**: Source-of-truth schemas, diagrams, and specifications.

| File | Required | Description |
|------|----------|-------------|
| `architecture.md` | yes | System architecture with ER diagram |
| `schema.md` | yes | Database schema documentation |
| `shipment-lifecycle.md` | yes | Shipment state machine diagram |
| `release.md` | yes | Release process and versioning |
| `glue.md` | no | Glue system documentation |
| `docs.md` | yes | This file (documentation schema) |

---

## Validation Rules

Rules that `/docs-doctor` checks against this schema.

### Directory Rules

1. **Required directories exist**: `docs/dev/`, `docs/guide/`, `docs/reference/`
2. **No top-level docs**: All `.md` files must be inside a subdirectory (no `docs/*.md`)
3. **No unexpected subdirectories**: Only `dev/`, `guide/`, `reference/` allowed under `docs/`

### File Rules

4. **Required files present**: Every file marked `required: yes` above must exist
5. **No orphan files**: Every file in `docs/` must be listed in this schema (required or optional)
6. **Glossary format**: `docs/guide/glossary.md` must contain term definitions, no mermaid blocks

### Content Rules

7. **Internal links valid**: Markdown links between docs files must resolve
8. **Lane separation**: Dev docs should not contain user-facing content; guide docs should not contain implementation details
