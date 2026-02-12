# Documentation Schema

Defines the structure and validation rules for ORC's documentation. Used by `/docs-doctor` to validate the docs directory.

---

## Directory Structure

```
docs/
├── README.md              # This file (documentation schema)
├── architecture.md        # System architecture with ER diagram
├── common-workflows.md    # Day-to-day workflow recipes
├── getting-started.md     # First-run setup and orientation
├── schema.md              # Database schema and glossary
├── tmux.md                # TMux session management guide
├── troubleshooting.md     # Common issues and fixes
└── dev/                   # Agent-facing: developing ORC itself
    ├── checklists.md
    ├── config.md
    ├── database.md
    ├── deployment.md
    ├── glue.md
    ├── git-hooks.md
    ├── integration-tests.md
    ├── release.md
    └── testing.md
```

### docs/ (top-level) — General Documentation

**Audience**: Both humans and agents.

**Purpose**: Getting started, workflows, reference schemas, and architecture.

| File | Required | Description |
|------|----------|-------------|
| `README.md` | yes | This file (documentation schema and validation rules) |
| `architecture.md` | yes | System architecture with ER diagram |
| `common-workflows.md` | yes | Day-to-day workflow recipes |
| `getting-started.md` | yes | First-run setup and orientation |
| `schema.md` | yes | Database schema, glossary, and terminology |
| `troubleshooting.md` | yes | Common issues and fixes |
| `tmux.md` | no | TMux session management guide |

### docs/dev/ — Development Documentation

**Audience**: Agents (IMPs, Goblins) developing ORC code.

**Purpose**: How to build, deploy, and modify ORC internals.

| File | Required | Description |
|------|----------|-------------|
| `checklists.md` | yes | Add Field, Add Entity, Add CLI, Add State checklists |
| `config.md` | yes | Config format, infrastructure plan/apply |
| `database.md` | yes | Atlas workflow, two-database model, schema changes |
| `deployment.md` | yes | Deployment checklist and merge-to-master process |
| `testing.md` | yes | Table-driven tests, test pyramid, verification discipline |
| `release.md` | yes | Release process and versioning |
| `git-hooks.md` | no | Git hook enforcement: what each hook checks and why |
| `glue.md` | no | Glue system documentation |
| `integration-tests.md` | no | Integration test skills: coverage map, gaps, and when to run |

---

## Validation Rules

Rules that `/docs-doctor` checks against this schema.

### Directory Rules

1. **One subdirectory only**: Only `dev/` is allowed under `docs/`
2. **Top-level .md files allowed**: General documentation lives directly in `docs/`

### File Rules

3. **Required files present**: Every file marked `required: yes` above must exist
4. **No orphan files**: Every file in `docs/` must be listed in this schema (required or optional)

### Content Rules

5. **Internal links valid**: Markdown links between docs files must resolve
6. **Lane separation**: Dev docs should not contain user-facing content; general docs should not contain implementation details
