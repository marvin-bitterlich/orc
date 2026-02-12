# Git Hooks

What each git hook enforces and why. Hooks are installed via `make init` (which runs `make install-hooks`). Source files live in `scripts/hooks/`.

---

## Pre-commit Hook (Blocking)

The pre-commit hook is a quality gate. Commits fail if any check does not pass.

### CHANGELOG Check

On feature branches (anything other than `master`/`main`), the hook diffs `CHANGELOG.md` against `master`. If there are no changes, the commit is rejected. This ensures every feature branch documents its changes in the `[Unreleased]` section.

### `make lint`

The lint target runs six sub-checks in sequence. All must pass.

#### schema-check

Prevents hardcoded `CREATE TABLE IF NOT EXISTS` statements in SQLite test files (`internal/adapters/sqlite/*_test.go`). Tests must use `db.GetSchemaSQL()` instead. This prevents schema drift where tests pass against a stale schema but production queries fail against the real one.

Also verifies that `internal/adapters/sqlite/testutil_test.go` calls `db.GetSchemaSQL()`.

| Config | Location |
|--------|----------|
| Inline in Makefile | `Makefile` (lines 112-125) |

#### check-test-presence

Ensures that certain source files have corresponding test files:

| Source pattern | Required test file |
|----------------|-------------------|
| `internal/app/*_service.go` | `internal/app/*_service_test.go` |
| `internal/adapters/sqlite/*_repo.go` | `internal/adapters/sqlite/*_repo_test.go` |
| `internal/core/*/guards.go` | `internal/core/*/guards_test.go` |

Files can be exempted by adding them to `scripts/test-presence-allowlist.txt` (one path per line, exact match).

| Config | Location |
|--------|----------|
| Check script | `scripts/check-test-presence.sh` |
| Allowlist | `scripts/test-presence-allowlist.txt` |

#### check-coverage

Runs `go test -cover` per package and enforces minimum coverage thresholds:

| Package | Threshold |
|---------|-----------|
| `internal/core/*` | 70% |
| `internal/app` | 50% |
| `internal/adapters/sqlite` | 60% |
| `internal/adapters/filesystem` | 50% |

Exempt packages (skipped silently): `internal/core/effects`, `internal/core/factory`, `internal/core/git`, `internal/core/workbench`, `internal/core/workshop`, `internal/cli`, `internal/tmux`.

Packages with 0% coverage and no test failures are also skipped (no statements to cover).

| Config | Location |
|--------|----------|
| Thresholds and exemptions | `scripts/check-coverage.sh` |

#### check-skills

Validates skill definitions in `glue/skills/` and `.claude/skills/`:

1. Every skill directory must contain a `SKILL.md` file
2. Every `SKILL.md` must have YAML frontmatter with `name:` and `description:` fields
3. Every skill on disk must appear in `docs/architecture.md` (skill table)
4. Every skill listed in `docs/architecture.md` must exist on disk

| Config | Location |
|--------|----------|
| Check script | `scripts/check-skills.sh` |
| Skill source dirs | `glue/skills/`, `.claude/skills/` |
| Skill registry | `docs/architecture.md` |

#### golangci-lint

Runs `golangci-lint run ./...`. Key enabled linters:

| Linter | Purpose |
|--------|---------|
| `govet` | Suspicious constructs (all checks enabled except `fieldalignment`, `shadow`) |
| `staticcheck` | Comprehensive static analysis (all checks) |
| `unused` | Unused code detection (disabled on test files) |
| `gosimple` | Simplification suggestions |
| `ineffassign` | Ineffectual assignments |
| `gofmt` | Formatting |
| `goimports` | Import ordering (local prefix: `github.com/example/orc`) |
| `depguard` | Import restrictions (see below) |
| `misspell` | Misspelled words (US locale) |
| `durationcheck` | Duration multiplication issues |
| `nilerr` | Incorrect nil error returns |
| `predeclared` | Shadowed predeclared identifiers |
| `unconvert` | Unnecessary type conversions |

**depguard rule** -- `core-no-exec`: Files in `internal/core/**/*.go` cannot import `os/exec`. Core must be pure; use the effects pattern and adapters for I/O.

| Config | Location |
|--------|----------|
| Linter config | `.golangci.yml` |

#### go-arch-lint

Runs `go-arch-lint check`. Enforces hexagonal architecture layer boundaries via import direction rules. Inner layers must not depend on outer layers.

Key dependency rules:

| Component | May depend on |
|-----------|--------------|
| `core` | `core` only (no project deps, no vendor deps) |
| `ports` | `ports` only |
| `models` | `models` only |
| `app` | `core`, `ports`, `models`, `config`, `context`, `agent` |
| `adapters` | `adapters`, `ports`, `models`, `db`, `core`, `config`, `context`, `ctxutil`, `agent`, `tmux` |
| `wire` | `app`, `adapters`, `ports`, `models`, `db`, `config`, `context`, `agent`, `tmux`, `version` |
| `cli` | `wire`, `ports`, `models`, `config`, `context`, `version`, `templates`, `scaffold`, `agent`, `db` |
| `cmd` | `cli`, `version` |

| Config | Location |
|--------|----------|
| Architecture rules | `.go-arch-lint.yml` |

### `make test`

After lint passes, the pre-commit hook runs `go test ./...` (the full test suite).

---

## Post-checkout Hook (Informational)

Runs after `git checkout` on branch switches only (not file checkouts). Does not block.

**What it does:**

1. Prints the current branch name
2. If Atlas is installed, checks for schema drift between the database and `schema.sql`
   - Suggests `make schema-diff` (preview) and `make schema-apply` (apply) if drift is detected
3. If Atlas is not installed, suggests installing it via `brew install ariga/tap/atlas`
4. If `.orc/workbench.db` exists, prints a reminder about the two-database model (`orc-dev` vs `orc`)

---

## Post-merge Hook (Informational)

Runs after `git merge` (including `git pull`). Does not block.

**What it does:**

1. Schema drift check (same as post-checkout)
2. On `master`/`main` branch, prints a full checklist:
   - `make init` -- Initialize dependencies
   - `make install` -- Install orc binary
   - `make deploy-glue` -- Deploy Claude Code glue
   - `make test` -- Verify tests pass
3. Runs `orc doctor --quiet` for environment health check (if `orc` is installed)
4. Release freshness hint: if a `VERSION` file exists and commits have landed since the last version tag, prints the commit count, days since tag, and suggests running `/release`

---

## Debugging Hook Failures

### Common pre-commit failures

| Symptom | Cause | Fix |
|---------|-------|-----|
| "CHANGELOG.md has no changes vs master" | Feature branch missing changelog entry | Add changes to `[Unreleased]` section in `CHANGELOG.md` |
| "Found hardcoded CREATE TABLE in test files" | Test file has inline schema | Replace with `db.GetSchemaSQL()` call |
| "MISSING: internal/app/foo_service_test.go" | New service file without test | Create the test file, or add to `scripts/test-presence-allowlist.txt` |
| "FAIL: internal/core/foo: 45.0% < 70%" | Coverage below threshold | Add tests to raise coverage above the threshold |
| "FRONTMATTER: skill-name missing 'name:' field" | Skill SKILL.md missing frontmatter | Add `name:` and `description:` to YAML frontmatter |
| "UNDOCUMENTED: skill-name exists but not in ARCHITECTURE.md" | New skill not registered | Add skill to the table in `docs/architecture.md` |
| depguard error in core package | Core code imports `os/exec` | Use effects pattern and adapters for I/O instead |
| go-arch-lint violation | Import crosses layer boundary | Move code to the correct layer or use a port |

### The `--no-verify` escape hatch

```bash
git commit --no-verify
```

Bypasses all pre-commit checks. Use only in emergencies (e.g., time-critical hotfix where you will fix lint issues in a follow-up). This is audited -- the team lead will review commits made with `--no-verify`.

---

## Config Locations

| Check | Config file(s) |
|-------|----------------|
| CHANGELOG enforcement | `scripts/hooks/pre-commit` (inline) |
| schema-check | `Makefile` (inline) |
| check-test-presence | `scripts/check-test-presence.sh`, `scripts/test-presence-allowlist.txt` |
| check-coverage | `scripts/check-coverage.sh` (thresholds and exemptions) |
| check-skills | `scripts/check-skills.sh` |
| golangci-lint | `.golangci.yml` |
| go-arch-lint | `.go-arch-lint.yml` |
| Hook installation | `Makefile` (`install-hooks` target) |
| Hook source files | `scripts/hooks/pre-commit`, `scripts/hooks/post-checkout`, `scripts/hooks/post-merge` |
