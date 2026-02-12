---
name: ship-deploy
description: |
  Deploy shipment by merging worktree branch to main repository, or by creating
  a PR against upstream when the repo is in fork mode.
  Use when user says "/ship-deploy", "ship deploy", "deploy shipment", "merge to master".
  Dynamically discovers deployment workflow from repo documentation.
next: ship-complete
---

# Ship Deploy

Deploy the current worktree branch by merging to the main repository, following the repo's deployment workflow.

## Workflow Discovery

This skill dynamically discovers how to deploy based on repo documentation:

### 1. Find Deployment Docs

Check for deployment documentation in order:
1. `docs/dev/deployment.md`
2. `DEPLOYMENT.md`
3. `CLAUDE.md` (look for deployment section)

If found, parse for:
- **Workflow type**: `direct` (merge to master), `pr` (pull request), or `fork` (PR against upstream)
- **Host repo path**: Where the main repo lives (for direct workflow)
- **Pre-merge checks**: Commands to run before merging
- **Post-merge steps**: Commands to run after merging

### 2. Detect Build System

Infer test/lint commands from build system:

| Build System | Test Command | Lint Command |
|--------------|--------------|--------------|
| Makefile | `make test` | `make lint` |
| package.json | `npm test` | `npm run lint` |
| Gemfile | `bundle exec rspec` | `bundle exec rubocop` |

If no build system found, warn and skip automated checks.

### 3. Check Fork Mode

```bash
orc repo show <repo-id>
```

Check if the repo has an upstream URL configured. If `upstream_url` is present, the repo is in fork mode regardless of deployment doc workflow type.

### 4. Determine Workflow

**If `orc repo show` returns an upstream URL:**
- Follow fork workflow (push to fork, PR against upstream)

**If deployment docs found with `type: direct`:**
- Follow direct-to-master workflow (merge, post-merge steps, push)

**If deployment docs found with `type: pr` or no docs found:**
- Fall back to PR workflow

## Direct Workflow

When deployment docs specify direct workflow:

### 1. Pre-flight Checks

```bash
# Must be clean
git status --porcelain  # Should be empty

# Run pre-merge checks from deployment docs
# Or infer from build system
```

### 2. Merge to Main Repo

```bash
cd <host-repo-path>  # From deployment docs
git checkout <main-branch>
git pull origin <main-branch>
git merge <worktree-branch> --no-edit
```

### 3. Post-Merge Steps

Run commands from deployment docs (e.g., `make init`, `make install`, etc.)

### 4. Push and Rebase

```bash
git push origin <main-branch>
cd <worktree-path>
git rebase <main-branch>
```

### 5. Note Deployment

The shipment remains in "in-progress" status. Deployment is tracked by the merge to master, not a separate status transition.

## Fork Workflow

When `orc repo show` returns an upstream URL, the repo is a fork. Deployment creates a PR against the upstream repository instead of merging locally.

### 1. Pre-flight Checks

```bash
# Must be clean
git status --porcelain  # Should be empty

# Run pre-merge checks from deployment docs
# Or infer from build system
```

### 2. Detect Fork and Upstream Info

```bash
orc repo show <repo-id>
```

Extract from output:
- **upstream_url**: The canonical repo URL (e.g. `git@github.com:owner/repo.git`)
- **upstream_branch**: The target branch on upstream (defaults to repo's `default_branch`)

Parse owner/repo from the upstream URL:

```bash
# SSH format: git@github.com:owner/repo.git → owner/repo
# HTTPS format: https://github.com/owner/repo.git → owner/repo
```

Also determine the fork owner from origin:

```bash
# Get fork owner from origin remote
git remote get-url origin
# Parse owner from origin URL using same SSH/HTTPS parsing
```

### 3. Push Branch to Fork

```bash
git push origin <branch>
```

### 4. Create PR Against Upstream

```bash
gh pr create \
  --repo <upstream-owner/repo> \
  --head <fork-owner>:<branch> \
  --base <upstream-branch> \
  --fill
```

### 5. Note Deployment

The shipment remains in "in-progress" status. Local merge and post-merge steps are **skipped** — those happen after the upstream repository merges the PR. Run `/ship-freshness` after the PR is merged to sync your fork.

## PR Workflow (Fallback)

When no deployment docs found or workflow type is `pr`:

### 1. Pre-flight Checks

```bash
git status --porcelain  # Must be clean
```

Run any discovered test/lint commands (warn if none found).

### 2. Create Pull Request

Check for PR creation skill first, then fall back to gh CLI:

```bash
# If gh CLI available
gh pr create --fill
```

### 3. Output

```
No deployment docs found. Using PR workflow.

Created PR: <url>

Follow your repository's review and merge process.
```

## Build System Detection

```bash
# Check for build systems in order
if [ -f "Makefile" ]; then
    TEST_CMD="make test"
    LINT_CMD="make lint"
elif [ -f "package.json" ]; then
    TEST_CMD="npm test"
    LINT_CMD="npm run lint"
elif [ -f "Gemfile" ]; then
    TEST_CMD="bundle exec rspec"
    LINT_CMD="bundle exec rubocop"
else
    echo "Warning: No build system detected. Skipping automated checks."
fi
```

## Success Output

**Direct workflow:**
```
Deployed via direct workflow:
  - Branch merged to <main-branch>
  - Post-merge steps completed
  - Pushed to origin
  - Worktree rebased
  - Ready to close shipment

Run /ship-complete to close the shipment.
```

**Fork workflow:**
```
Deployed via fork workflow:
  - Branch pushed to origin (fork)
  - PR created against upstream: <url>
  - Targeting <upstream-owner/repo> branch <upstream-branch>
  - After PR merges, run /ship-freshness to sync fork
  - Then run /ship-complete to close the shipment
```

**PR workflow:**
```
Created pull request:
  - PR: <url>
  - Follow your repo's review process
  - After PR merges, run /ship-complete to close the shipment
```

## Error Handling

| Error | Action |
|-------|--------|
| Dirty working tree | List uncommitted files, ask to commit or stash |
| Pre-merge checks fail | Show failures, do not proceed |
| Merge conflicts | Report conflicts, do not auto-resolve |
| No gh CLI for PR | Provide manual instructions |
| Push rejected | Check if remote has new commits |
| Upstream URL not parseable | Show URL, ask user to verify `orc repo show` output |
| gh CLI can't reach upstream | Verify upstream URL and permissions, suggest checking `gh auth status` |
| Fork owner detection fails | Show origin URL, ask user to verify `git remote get-url origin` |
