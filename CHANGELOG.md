# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **Documentation restructure**: CLAUDE.md slimmed from ~800 lines to ~80 lines (constitution only); content extracted to docs/dev/ (checklists, database, testing, config). Docs reorganized into docs/dev/ (for agents), docs/guide/ (for humans), docs/reference/ (shared). All cross-references updated.
- Shipment `status` CLI command no longer shows escape hatch warning (all transitions are manual now)
- Removed dead NudgeSession code from tmux port, adapter, and implementation
- Cleaned stale comments referencing gatehouse, approval, escalation, and other removed entities
- Updated docs (architecture.md, schema.md, glossary.md, README.md, CLAUDE.md) to remove approval/escalation/handoff references
- Fixed orc-self-test skill to remove gatehouse references
- **Actor model redesign**: Goblin is now the coordinator (human's workbench pane), IMP is a disposable worker (spawned by Claude Teams)
- Shipment lifecycle simplified to 4 statuses: draft → ready → in-progress → closed (all transitions manual)
- Task lifecycle simplified to 4 statuses: open → in-progress → closed (blocked as lateral state)
- `orc connect` retains `--role` flag but removes place-based autodetection
- **Skill reorganization**: Dev-only skills (bootstrap-exercise, release, orc-self-test, orc-architecture) moved from glue/ to .claude/skills/ (repo-local)
- ship-plan now includes dependency analysis guidance for task parallelism
- Summary display: tomes now expand when their parent entity (shipment or commission) is focused
- `deploy-glue` Makefile target now cleans orphan skill copies from ~/.claude/skills/

### Removed

- Approval entity: table, indexes, repo, service, tests, CLI command, port, wire
- Escalation entity: table, indexes, repo, service, tests, CLI command, port, wire
- Handoff entity: table, index, repo, service, tests, CLI command, port, model, wire
- Plan lifecycle simplified: statuses narrowed from 5 (draft, pending_review, approved, escalated, superseded) to 2 (draft, approved); removed SubmitPlan, EscalatePlan operations, supersedes_plan_id column, and related guards
- Gatehouse entity, GATE- place IDs, and all gatehouse infrastructure (absorbed into workbench)
- Message/mail system (replaced by Claude Teams messaging)
- Receipt entity and all receipt CLI/service/repo/guards
- Nudge CLI command (autorun propulsion removed)
- Auto-transition machinery (GetAutoTransitionStatus, AutoTransitionContext, event triggers)
- 9 shipment statuses collapsed: exploring, synthesizing, specced, planned, tasked, ready_for_imp, implementing, auto_implementing, implemented, verified, deployed
- 2 task statuses removed: ready, paused
- CLI commands: shipment auto/manual/ready/should-continue, shipment deploy/verify
- Skills deleted: imp-poll, imp-start, imp-implement, imp-auto, imp-nudge, imp-respawn, imp-rec, imp-plan-create, imp-plan-submit, goblin-escalation-receive, watchdog-monitor
- Watchdog state model removed: kennel, patrol, dogbed, stuck, check entities and all related CLI commands, skills, and documentation
- Skills deleted: orc-ping, ship-queue, imp-escalate
- Conclave entity and all references (replaced by shipment)
- Shipyard entity and all references

### Added

- `--depends-on` flag on `orc task create` for expressing task dependency relationships
- `ship-run` skill: bridges ship-plan → Claude Teams execution (dependency graph, team shape proposal, preflight checks, context injection)
- `orc backfill lifecycle-statuses` command for migrating existing data to new status values
- `closed_reason` column on shipments (nullable)

### Fixed

- `orc infra apply` now handles partial apply recovery by falling back to session name detection when ORC_WORKSHOP_ID env var is missing
- Bootstrap test now creates `~/.claude/settings.json` stub before `make bootstrap` to prevent jq merge failures on fresh VMs
- `/bootstrap-exercise` skill cleanup instructions now use archive + infra apply pattern instead of deprecated delete commands

### Added

- Watchdog monitoring system reimplemented:
  - `orc shipment should-continue` command for checking if IMP should continue working
  - Watchdog pane (index 4) added to infra plan/apply when patrol is active
  - `orc prime` detects WATCHDOG_KENNEL_ID env var and provides watchdog context
  - `/watchdog-monitor` skill for monitoring loop (check continue, capture pane, detect state, take action)
  - `/imp-auto` skill now spawns watchdog via patrol start + infra apply
  - `/imp-start --auto` integrates watchdog spawning flow
  - `/orc-self-test` includes comprehensive watchdog infrastructure tests
- `make bootstrap` now creates REPO-001 (ORC repository at ~/src/orc) after FACT-001
- `orc bootstrap --factory FACT-xxx` flag passes factory to /orc-first-run skill
- `/bootstrap-exercise` skill for manual testing of first-run flow with isolated factory
- Bootstrap test verifies FACT-001 and REPO-001 creation
- `Brewfile` and `Brewfile.dev` for Homebrew dependency management
- `make bootstrap-dev` target for installing development dependencies (tart, sshpass, atlas)
- `--strict` flag for `orc doctor` - treats warnings as errors (useful for CI/scripts)
- Directory guards on dangerous Make targets (install, deploy-glue, schema-apply, bootstrap)
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

- `/orc-first-run` skill uses canonical path ~/src/orc instead of $(pwd), accepts factory from directive
- `docs/getting-started.md` recommends `orc bootstrap` as step 3 after make bootstrap
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
