# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Bootstrap test now creates `~/.claude/settings.json` stub before `make bootstrap` to prevent jq merge failures on fresh VMs

### Added

- `make bootstrap` now creates REPO-001 (ORC repository at ~/src/orc) after FACT-001
- `orc bootstrap --factory FACT-xxx` flag passes factory to /orc-first-run skill
- `/bootstrap-exercise` skill for manual testing of first-run flow with isolated factory
- Bootstrap test verifies FACT-001 and REPO-001 creation

### Changed

- `/orc-first-run` skill uses canonical path ~/src/orc instead of $(pwd), accepts factory from directive
- `docs/getting-started.md` recommends `orc bootstrap` as step 3 after make bootstrap

- `Brewfile` and `Brewfile.dev` for Homebrew dependency management
- `make bootstrap-dev` target for installing development dependencies (tart, sshpass, atlas)
- `--strict` flag for `orc doctor` - treats warnings as errors (useful for CI/scripts)
- Directory guards on dangerous Make targets (install, deploy-glue, schema-apply, bootstrap)
- SubagentStop hook support: `orc hook SubagentStop` complements Stop hook by catching skill/subagent completion points
- `--shell` flag for bootstrap-test: drops into interactive VM shell after bootstrap (`make bootstrap-shell`)
- `--keep` flag for bootstrap-test: preserves VM for manual exploration
- Bootstrap VM testing with Tart: `make bootstrap-test` spins up fresh macOS VM to validate first-run experience
- `/bootstrap-test` skill for running VM-based bootstrap validation
- `/release` skill now runs `/bootstrap-test` after docs-doctor (hard blocker)
- Bootstrap VM Testing documentation in CLAUDE.md
- `orc bootstrap` CLI command - starts Claude with `/orc-first-run` skill for first-time setup
- `/release` skill now runs `/docs-doctor` validation before release (hard blocker)
- Pre-commit hook enforces CHANGELOG.md changes on feature branches
- Post-merge hook runs `orc doctor` on master/main branch
- `/docs-doctor` skill checks for repo-agnosticism violations in skills
- Guardrail enforcement documentation in CLAUDE.md
- Self-test skill now verifies tmux session management
- `docs/schema.md` - dedicated ER diagram with 12 core tables
- `docs/shipment-lifecycle.md` - dedicated shipment state machine diagram

### Changed

- **BREAKING**: ORC must be cloned to `~/src/orc` (canonical location enforced by orc doctor)
- Bootstrap now requires Homebrew and runs `brew bundle` to install Go
- Bootstrap test copies repo to `~/src/orc` in VM and runs `orc doctor` verification
- Factory case mismatch fixed: Makefile now creates `default` (lowercase) to match service lookup
- Improved `orc doctor` messaging: "All checks passed" only shown when truly clean, warnings now properly distinguished
- Enhanced repo check validates both directory existence and `.git` presence
- Bootstrap test now validates CLI functionality (creates commission, workshop, runs summary)
- `/orc-first-run` skill rewritten for adaptive onboarding - checks existing state, creates missing entities, guides repo/template setup
- Makefile bootstrap target now points to `orc bootstrap` instead of `/orc-first-run`
- Simplified `docs/shipment-lifecycle.md` to two phases: Planning and Implementation
- Consolidated `docs/glossary/` into single `docs/glossary.md` (A-Z term list)
- Stripped `docs/troubleshooting.md` to only `/orc-help` and `orc doctor`
- `docs-doctor` skill now validates glossary structure (no mermaid, no outbound links)
- Fixed `--repo` → `--repo-id` flag in `docs/getting-started.md`
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

- `docs-doctor` glossary check no longer fails on outbound links (cross-references allowed)

### Security
