---
name: release
description: |
  Cut a new release by bumping VERSION, promoting CHANGELOG, committing, and tagging.
  Use when user says "/release", "cut a release", "bump version", or "create release".
---

# Release

Cut a new semantic version release by bumping VERSION, promoting CHANGELOG Unreleased section, creating a release commit, and tagging.

## Usage

```
/release           (suggest patch, ask to confirm)
/release --patch   (bump patch: 0.1.0 → 0.1.1)
/release --minor   (bump minor: 0.1.0 → 0.2.0)
/release --major   (bump major: 0.1.0 → 1.0.0)
```

## Flow

### Step 1: Pre-flight Checks

```bash
git status --porcelain
```

Must be clean. If dirty:
```
❌ Working tree is dirty. Commit or stash changes first.
```

### Step 2: Validate Documentation

Run docs-doctor to ensure documentation is current before release:

```
/docs-doctor
```

If docs-doctor fails:
```
❌ Documentation validation failed. Fix issues before releasing.

Run /docs-doctor --fix to auto-fix simple issues, then retry.
```

This is a hard blocker - releases cannot proceed with invalid documentation.

### Step 3: Validate Bootstrap

Run bootstrap-test to ensure `make bootstrap` works for new users:

```
/bootstrap-test
```

If bootstrap-test fails:
```
❌ Bootstrap test failed. Fix issues before releasing.

Run /bootstrap-test --keep-on-failure --verbose to debug.
```

This is a hard blocker - releases cannot proceed if bootstrap is broken.

### Step 4: Read Current Version

```bash
cat VERSION
```

Parse current version (e.g., `0.1.0`).

If VERSION file missing:
```
❌ No VERSION file found. Create one first.
```

### Step 5: Check CHANGELOG

```bash
cat CHANGELOG.md
```

Look for content under `## [Unreleased]` section.

If Unreleased section is empty (only headers, no content):
```
⚠️ CHANGELOG Unreleased section is empty.

Proceed anyway? [y/n]
```

Allow proceeding if user confirms.

### Step 6: Determine Version Bump

**If flag provided:**
- `--patch`: increment patch (0.1.0 → 0.1.1)
- `--minor`: increment minor, reset patch (0.1.0 → 0.2.0)
- `--major`: increment major, reset minor and patch (0.1.0 → 1.0.0)

**If no flag:**
```
Current version: 0.1.0

Suggested bump: patch → 0.1.1

Accept? [y/n/minor/major]
```

- `y` or Enter: use patch
- `n`: abort
- `minor`: bump minor instead
- `major`: bump major instead

### Step 7: Update VERSION File

Write new version to VERSION file:
```bash
echo "0.1.1" > VERSION
```

### Step 8: Update CHANGELOG

Replace `## [Unreleased]` section with versioned section:

**Before:**
```markdown
## [Unreleased]

### Added
- New feature X
```

**After:**
```markdown
## [Unreleased]

## [0.1.1] - 2026-02-08

### Added
- New feature X
```

Use today's date in ISO format (YYYY-MM-DD).

### Step 9: Create Release Commit

```bash
git add VERSION CHANGELOG.md
git commit -m "release: v0.1.1"
```

### Step 10: Create Tag

```bash
git tag v0.1.1
```

### Step 11: Push (Optional)

```
Release created locally:
  - VERSION: 0.1.1
  - Commit: release: v0.1.1
  - Tag: v0.1.1

Push to origin? [y/n]
```

If yes:
```bash
git push origin HEAD
git push origin v0.1.1
```

### Step 12: Success Output

```
✓ Released v0.1.1

  VERSION: 0.1.1
  Commit: <sha>
  Tag: v0.1.1
  Pushed: yes/no

Next:
  - Verify CI passes
  - Update any dependent systems
```

## Error Handling

| Error | Action |
|-------|--------|
| Dirty working tree | "Commit or stash changes first" |
| Docs-doctor fails | "Fix documentation issues before releasing" |
| Bootstrap-test fails | "Fix bootstrap issues before releasing" |
| No VERSION file | "Create VERSION file first" |
| Invalid version format | "VERSION must be semver (X.Y.Z)" |
| Tag already exists | "Tag vX.Y.Z already exists" |
| Push rejected | "Remote has new commits. Pull first." |

## Notes

- This is a repo-level skill, not a glue-level skill
- Tags use `v` prefix (v0.1.0)
- CHANGELOG follows keepachangelog.com format
- Empty Unreleased is a warning, not a blocker
- Docs-doctor validation is a hard blocker (no exceptions)
- Bootstrap-test validation is a hard blocker (no exceptions)
