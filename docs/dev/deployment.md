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
| Lint | `make lint` | Yes |
| Tests | `make test` | Yes |

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

## Post-Deploy Verification

After deployment, verify the installation works correctly.

### Verification Checks

Run these checks to confirm deployment success:

| Check | Command | Purpose |
|-------|---------|---------|
| 1 | `orc status` | Verify ORC is responsive |
| 2 | `orc commission list` | Confirm DB connectivity |
| 3 | `orc shipment list` | Validate data access |
| 4 | `make test` | Full test suite passes |

### Verification Results

All checks should pass:

```
Verification Results:
  [PASS] orc status
  [PASS] orc commission list
  [PASS] orc shipment list
  [PASS] make test
```

### Next Steps

After verification passes:

```bash
/ship-complete SHIP-XXX         # Close shipment (terminal state)
```

---

## Fork-Based Deployment

When a repo has `upstream_url` configured, it operates in **fork mode**. Instead of merging directly to master, changes are submitted as pull requests against the upstream (canonical) repository.

Fork mode is detected automatically: if `orc repo show` returns an `upstream_url`, the `/ship-deploy` skill uses the fork workflow. The direct workflow sections above still apply to non-fork repos.

### Fork Setup

1. Configure the repo for fork mode:
   ```bash
   orc repo fork REPO-xxx --upstream-url <url> [--upstream-branch main]
   ```

2. Verify the configuration:
   ```bash
   orc repo show REPO-xxx
   ```
   Output should include `Upstream URL` and `Upstream Branch` fields.

3. Ensure the `upstream` git remote exists locally:
   ```bash
   git remote add upstream <upstream-url>
   git fetch upstream
   ```

### Fork Deploy Flow

Pre-flight checks are the same as the direct workflow:

| Check | Command | Required |
|-------|---------|----------|
| Clean working tree | `git status --porcelain` | Yes |
| Lint | `make lint` | Yes |
| Tests | `make test` | Yes |

Then deploy via PR against upstream:

1. Push the worktree branch to your fork:
   ```bash
   git push origin <branch>
   ```

2. Create a PR targeting the upstream repository:
   ```bash
   gh pr create \
     --repo <upstream-owner/repo> \
     --head <fork-owner>:<branch> \
     --base <upstream-branch> \
     --fill
   ```

3. URL parsing — owner/repo is extracted from remote URLs:
   - **SSH**: `git@github.com:owner/repo.git` → `owner/repo`
   - **HTTPS**: `https://github.com/owner/repo.git` → `owner/repo`

4. There are **no local merge or post-merge steps**. Those happen after the upstream repository merges the PR.

### Upstream Sync

After an upstream PR is merged (or periodically to stay current), sync your fork:

```bash
git fetch upstream
git checkout <default_branch>
git merge upstream/<upstream_branch>
git push origin <default_branch>
git checkout <worktree_branch>
git rebase <default_branch>
```

The `/ship-freshness` skill automates this — it detects fork mode and runs the upstream sync flow instead of the standard `git rebase origin/master`.

### Fork Lifecycle Example

Complete workflow for a fork-based shipment:

```
# 1. Setup (one-time)
orc repo fork REPO-xxx --upstream-url git@github.com:upstream-org/repo.git
git remote add upstream git@github.com:upstream-org/repo.git

# 2. Develop (normal shipment flow)
orc shipment create --repo REPO-xxx --name "Add feature"
# ... work on branch, commit changes ...

# 3. Deploy (creates upstream PR)
/ship-deploy
# → Runs pre-flight checks
# → Pushes branch to origin (your fork)
# → Creates PR against upstream-org/repo
# → Outputs PR URL

# 4. After upstream merges the PR
/ship-freshness
# → Fetches from upstream
# → Merges upstream changes into default branch
# → Pushes to fork origin
# → Rebases worktree branch

# 5. Complete
/ship-complete SHIP-xxx
```
