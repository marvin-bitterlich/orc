# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `/release` skill now runs `/docs-doctor` validation before release (hard blocker)
- Pre-commit hook enforces CHANGELOG.md changes on feature branches
- Post-merge hook runs `orc doctor` on master/main branch
- `/docs-doctor` skill checks for repo-agnosticism violations in skills
- Guardrail enforcement documentation in CLAUDE.md
- Self-test skill now verifies tmux session management
- `docs/schema.md` - dedicated ER diagram with 12 core tables
- `docs/shipment-lifecycle.md` - dedicated shipment state machine diagram

### Changed

- `docs-doctor` skill moved from `glue/skills/` to `.claude/skills/` (repo-local, not globally deployed)
- `docs-doctor` now uses subset validation for diagrams (simplified diagrams are intentional)
- ER diagram and lifecycle diagram moved from inline to dedicated docs
- `docs/architecture.md` now links to `docs/schema.md` instead of inline diagram
- `docs/common-workflows.md` now links to `docs/shipment-lifecycle.md` instead of inline diagram

### Removed

- `DOCS.md` - documentation index consolidated into docs-doctor skill
- `/ship-verify` skill - removed from glue/skills/ and docs
- 8 orphaned skills from `~/.claude/skills/`: aws-bug-report, conclave, exorcism, merge-to-master, ship-tidy, ship-verify, watchdog, white-smoke

### Changed

### Deprecated

### Removed

### Fixed

### Security
