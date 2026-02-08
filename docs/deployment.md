# ORC Deployment

ORC uses a direct-to-master deployment workflow. Changes are merged directly from worktree branches into master without pull requests.

## Workflow Type

```
type: direct
```

## Host Repository

The main ORC repository lives at:

```
path: ~/src/orc
branch: master
```

## Pre-Merge Checks

Before merging to master, run these checks in the worktree:

| Check | Command | Required |
|-------|---------|----------|
| Clean working tree | `git status --porcelain` | Yes |
| CHANGELOG updated | `git diff master -- CHANGELOG.md` | Yes |
| Lint | `make lint` | Yes |
| Tests | `make test` | Yes |

**CHANGELOG Requirement:**
- Any change to CHANGELOG.md is required before merging feature branches
- No skip flag - discipline over convenience
- Emergency escape hatch: `git commit --no-verify` (use of this flag is audited)

## Merge Process

1. Switch to host repo and update master:
   ```bash
   cd ~/src/orc
   git checkout master
   git pull origin master
   ```

2. Merge the worktree branch:
   ```bash
   git merge <branch-name> --no-edit
   ```

## Post-Merge Steps

After merging, run these commands in the host repo:

| Step | Command | Purpose |
|------|---------|---------|
| 1 | `make init` | Install git hooks |
| 2 | `make install` | Build and install orc binary |
| 3 | `make deploy-glue` | Deploy Claude Code skills/hooks |
| 4 | `make test` | Verify build works |

## Schema Sync

If schema changes were included, sync the database:

```bash
make schema-diff    # Preview changes
make schema-apply   # Apply to local DB
```

## Push and Rebase

1. Push master to origin:
   ```bash
   git push origin master
   ```

2. Return to worktree and rebase:
   ```bash
   cd <worktree-path>
   git rebase master
   ```

## Shipment Status

After successful deployment, update shipment status:

```bash
orc shipment deploy SHIP-XXX
```
