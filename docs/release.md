# Release Workflow

ORC uses semantic versioning with a VERSION file as the source of truth.

## Versioning Scheme

| Component | Location | Purpose |
|-----------|----------|---------|
| VERSION | `VERSION` | Current release version (e.g., `0.1.0`) |
| CHANGELOG | `CHANGELOG.md` | Human-readable change history |
| Git tags | `v0.1.0` | Mark release commits |

### Version Detection

The Makefile compares HEAD to the tag commit:
- **HEAD matches tag**: Version is `v0.1.0`
- **HEAD ahead of tag**: Version is `v0.1.0-dev`

This is injected at build time via LDFLAGS.

## CHANGELOG Format

CHANGELOG.md follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format:

```markdown
## [Unreleased]

### Added
- New feature description

### Changed
- Modification description

### Fixed
- Bug fix description

## [0.1.0] - 2026-02-08

### Added
- Initial release features
```

**Sections** (use only those that apply):
- **Added** - New features
- **Changed** - Changes to existing functionality
- **Deprecated** - Soon-to-be removed features
- **Removed** - Removed features
- **Fixed** - Bug fixes
- **Security** - Vulnerability fixes

## /release Skill

Cut a new release using the `/release` skill:

```
/release           # Suggest patch, ask to confirm
/release --patch   # Bump patch: 0.1.0 → 0.1.1
/release --minor   # Bump minor: 0.1.0 → 0.2.0
/release --major   # Bump major: 0.1.0 → 1.0.0
```

The skill:
1. Validates working tree is clean
2. Reads current VERSION
3. Warns if CHANGELOG Unreleased is empty
4. Bumps VERSION file
5. Promotes Unreleased to versioned section
6. Creates commit: `release: vX.X.X`
7. Creates tag: `vX.X.X`
8. Optionally pushes to origin

## Git Tag Conventions

- Tags use `v` prefix: `v0.1.0`, `v1.0.0`
- Tags point to release commits
- Annotated tags not required (lightweight OK)

## Post-Merge Hints

On master/main branch, the post-merge hook shows release hints:

```
📦 12 commits since v0.1.0 (5 days ago)
   Run /release to cut a new version
```

This gives visibility into accumulated changes without mandating releases.

## Release Process

### Standard Flow

1. Work accumulates, CHANGELOG Unreleased section grows
2. After deploying significant changes, consider a release
3. Run `/release` (or agent offers after `/ship-deploy`)
4. VERSION bumped, CHANGELOG promoted
5. Commit created with message `release: vX.X.X`
6. Tag `vX.X.X` created
7. Push commit and tag to origin

### First Release

For new projects without tags:
1. Create `VERSION` file with initial version (e.g., `0.1.0`)
2. Create `CHANGELOG.md` with Unreleased section
3. Run `/release` to create first tagged release

## Semver Guidelines

ORC follows [Semantic Versioning 2.0.0](https://semver.org/):

- **Major (1.0.0)**: Breaking changes
- **Minor (0.1.0)**: New features, backwards compatible
- **Patch (0.0.1)**: Bug fixes, backwards compatible

**Pre-1.0**: Major version zero means anything can change. The API is not stable. Start at 0.1.0 and increment minor for features.

**Move to 1.0.0** when:
- Used in production by others
- API stability matters
- Managing backwards compatibility
