.PHONY: install install-binary install-shim uninstall-shim dev build test lint lint-fix schema-check check-test-presence check-coverage init install-hooks clean help deploy-glue schema-diff schema-apply schema-inspect setup-workbench schema-diff-workbench schema-apply-workbench

# Go binary location (handles empty GOBIN)
GOBIN := $(shell go env GOPATH)/bin

# Version info (injected at build time)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
LDFLAGS := -X 'github.com/example/orc/internal/version.Version=$(VERSION)' \
           -X 'github.com/example/orc/internal/version.Commit=$(COMMIT)' \
           -X 'github.com/example/orc/internal/version.BuildTime=$(BUILD_TIME)'

# Default target
.DEFAULT_GOAL := help

#---------------------------------------------------------------------------
# Installation (global binary + shim)
#---------------------------------------------------------------------------

# Full install: binary + shim
install: install-binary install-shim install-dev-shim
	@echo ""
	@echo "✓ ORC installed with local-first shim"
	@echo "  Binary: $(GOBIN)/orc-bin"
	@echo "  Shim:   $(GOBIN)/orc"
	@echo ""
	@echo "Usage:"
	@echo "  In orc repo with ./orc  → uses local (shows '[using local ./orc]')"
	@echo "  Elsewhere               → uses global orc-bin"

# Install the actual binary as orc-bin
install-binary:
	@echo "Building and installing orc-bin..."
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/orc-bin ./cmd/orc

# Install the shim script as orc
install-shim:
	@echo "Installing shim..."
	@echo '#!/bin/bash' > $(GOBIN)/orc
	@echo '# ORC local-first shim - prefers ./orc when present' >> $(GOBIN)/orc
	@echo 'if [[ -x "./orc" ]]; then' >> $(GOBIN)/orc
	@echo '  echo "[using local ./orc]" >&2' >> $(GOBIN)/orc
	@echo '  exec ./orc "$$@"' >> $(GOBIN)/orc
	@echo 'else' >> $(GOBIN)/orc
	@echo '  exec "$$(dirname "$$0")/orc-bin" "$$@"' >> $(GOBIN)/orc
	@echo 'fi' >> $(GOBIN)/orc
	@chmod +x $(GOBIN)/orc

# Install orc-dev shim for development (requires workbench DB)
install-dev-shim:
	@echo "Installing orc-dev shim..."
	@echo '#!/bin/bash' > $(GOBIN)/orc-dev
	@echo '# ORC dev shim - requires workbench-local DB' >> $(GOBIN)/orc-dev
	@echo 'if [[ ! -f ".orc/workbench.db" ]]; then' >> $(GOBIN)/orc-dev
	@echo '  echo "Error: No workbench DB found at .orc/workbench.db" >&2' >> $(GOBIN)/orc-dev
	@echo '  echo "Run: make setup-workbench" >&2' >> $(GOBIN)/orc-dev
	@echo '  exit 1' >> $(GOBIN)/orc-dev
	@echo 'fi' >> $(GOBIN)/orc-dev
	@echo 'export ORC_DB_PATH=".orc/workbench.db"' >> $(GOBIN)/orc-dev
	@echo 'exec "$$(dirname "$$0")/orc-bin" "$$@"' >> $(GOBIN)/orc-dev
	@chmod +x $(GOBIN)/orc-dev
	@echo "✓ orc-dev installed (requires .orc/workbench.db)"

# Remove shim and restore direct binary access
uninstall-shim:
	@if [ -f "$(GOBIN)/orc-bin" ]; then \
		rm -f $(GOBIN)/orc; \
		mv $(GOBIN)/orc-bin $(GOBIN)/orc; \
		echo "✓ Shim removed, binary restored to $(GOBIN)/orc"; \
	else \
		echo "No orc-bin found, nothing to restore"; \
	fi

#---------------------------------------------------------------------------
# Development (local binary)
#---------------------------------------------------------------------------

# Build local binary for development (preferred command)
dev:
	@echo "Building local ./orc..."
	@go build -ldflags "$(LDFLAGS)" -o orc ./cmd/orc
	@echo "✓ Built ./orc (local development binary)"

# Alias for backwards compatibility
build: dev

#---------------------------------------------------------------------------
# Testing & Maintenance
#---------------------------------------------------------------------------

# Run all tests
test:
	go test ./...

#---------------------------------------------------------------------------
# Linting
#---------------------------------------------------------------------------

# Run all linters (golangci-lint + architecture + schema-check + test checks)
lint: schema-check check-test-presence check-coverage
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	@golangci-lint run ./...
	@echo "Running architecture lint..."
	@command -v go-arch-lint >/dev/null 2>&1 || { echo "go-arch-lint not installed. Run: go install github.com/fe3dback/go-arch-lint@latest"; exit 1; }
	@go-arch-lint check
	@echo "✓ All linters passed"

# Validate test schemas use the authoritative schema.go
# This prevents schema drift where tests pass but production queries fail
# Protection layers:
#   1. schema-check: Blocks hardcoded CREATE TABLE in tests
#   2. Tests use db.GetSchemaSQL(): SQLite fails if queries reference missing columns
#   3. CI runs both lint (includes schema-check) and test
schema-check:
	@echo "Checking for hardcoded test schemas..."
	@if grep -r "CREATE TABLE IF NOT EXISTS" internal/adapters/sqlite/*_test.go 2>/dev/null | grep -v "^Binary"; then \
		echo "ERROR: Found hardcoded CREATE TABLE in test files"; \
		echo "Tests should use db.GetSchemaSQL() instead"; \
		exit 1; \
	fi
	@echo "✓ No hardcoded test schemas found"
	@echo "Checking testutil uses authoritative schema..."
	@if ! grep -q 'db.GetSchemaSQL()' internal/adapters/sqlite/testutil_test.go; then \
		echo "ERROR: testutil_test.go must use db.GetSchemaSQL()"; \
		exit 1; \
	fi
	@echo "✓ Test setup uses authoritative schema"

# Check that all source files have corresponding test files
check-test-presence:
	@./scripts/check-test-presence.sh

# Check coverage thresholds per package
check-coverage:
	@./scripts/check-coverage.sh

# Run golangci-lint with auto-fix
lint-fix:
	@echo "Running golangci-lint with --fix..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	@golangci-lint run --fix ./...
	@echo "✓ Lint fixes applied"

#---------------------------------------------------------------------------
# Schema Management (Atlas)
#---------------------------------------------------------------------------

# Preview schema changes (diff current DB vs schema.sql)
schema-diff:
	@echo "Comparing current database to schema.sql..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema apply --env local --dry-run

# Apply schema changes from schema.sql to database
schema-apply:
	@echo "Applying schema.sql to database..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema apply --env local --auto-approve

# Dump current database schema
schema-inspect:
	@echo "Inspecting current database schema..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema inspect --env local

# Schema management for workbench-local database
schema-diff-workbench:
	@if [ ! -f ".orc/workbench.db" ]; then \
		echo "No workbench DB found. Run: make setup-workbench"; \
		exit 1; \
	fi
	@echo "Comparing workbench DB to schema.sql..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema apply --env workbench --dry-run

schema-apply-workbench:
	@if [ ! -f ".orc/workbench.db" ]; then \
		echo "No workbench DB found. Run: make setup-workbench"; \
		exit 1; \
	fi
	@echo "Applying schema.sql to workbench DB..."
	@command -v atlas >/dev/null 2>&1 || { echo "atlas not installed. Run: brew install ariga/tap/atlas"; exit 1; }
	atlas schema apply --env workbench --auto-approve

#---------------------------------------------------------------------------
# Development Environment Setup
#---------------------------------------------------------------------------

# Install git hooks (handles both regular repos and worktrees)
install-hooks:
	@HOOKS_DIR=$$(git rev-parse --git-common-dir)/hooks; \
	mkdir -p "$$HOOKS_DIR"; \
	cp scripts/hooks/pre-commit "$$HOOKS_DIR/pre-commit"; \
	chmod +x "$$HOOKS_DIR/pre-commit"; \
	cp scripts/hooks/post-merge "$$HOOKS_DIR/post-merge"; \
	chmod +x "$$HOOKS_DIR/post-merge"; \
	cp scripts/hooks/post-checkout "$$HOOKS_DIR/post-checkout"; \
	chmod +x "$$HOOKS_DIR/post-checkout"; \
	echo "✓ Git hooks installed to $$HOOKS_DIR"

# Initialize development environment
init: install-hooks
	@echo "✓ ORC development environment initialized"

# Setup workbench-local development database
setup-workbench:
	@echo "Creating workbench-local database..."
	@mkdir -p .orc
	@rm -f .orc/workbench.db
	@ORC_DB_PATH=.orc/workbench.db go run ./cmd/orc dev reset --force
	@echo ""
	@echo "✓ Workbench DB created: .orc/workbench.db"
	@echo ""
	@echo "Usage:"
	@echo "  orc-dev ...    → uses this local DB (when present)"
	@echo "  orc ...        → uses production DB"

# Clean build artifacts
clean:
	rm -f orc
	go clean
	@echo "✓ Cleaned local build artifacts"

#---------------------------------------------------------------------------
# Claude Code Integration (Glue)
#---------------------------------------------------------------------------

# Deploy skills and hooks to Claude Code
deploy-glue:
	@echo "Deploying Claude Code skills..."
	@for dir in glue/skills/*/; do \
		name=$$(basename "$$dir"); \
		echo "  → $$name"; \
		rm -rf ~/.claude/skills/$$name; \
		cp -r "$$dir" ~/.claude/skills/$$name; \
	done
	@echo "✓ Skills deployed to ~/.claude/skills/"
	@if [ -d "glue/hooks" ] && [ "$$(ls -A glue/hooks 2>/dev/null)" ]; then \
		echo "Deploying Claude Code hooks..."; \
		for hook in glue/hooks/*.sh; do \
			[ -f "$$hook" ] || continue; \
			name=$$(basename "$$hook"); \
			echo "  → $$name"; \
			cp "$$hook" ~/.claude/hooks/$$name; \
			chmod +x ~/.claude/hooks/$$name; \
		done; \
		echo "✓ Hooks deployed to ~/.claude/hooks/"; \
		if [ -f "glue/hooks.json" ]; then \
			echo "Configuring hooks in settings.json..."; \
			jq -s '.[0].hooks = (.[0].hooks // {}) * .[1] | .[0]' \
				~/.claude/settings.json glue/hooks.json > /tmp/settings.json && \
				mv /tmp/settings.json ~/.claude/settings.json; \
			echo "✓ Hooks configured in settings.json"; \
		fi; \
	fi

#---------------------------------------------------------------------------
# Help
#---------------------------------------------------------------------------

help:
	@echo "ORC Makefile Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev           Build local ./orc for development"
	@echo "  make test          Run all tests"
	@echo "  make lint          Run golangci-lint + architecture + schema-check"
	@echo "  make lint-fix      Run golangci-lint with auto-fix"
	@echo "  make schema-check  Verify test files use authoritative schema"
	@echo "  make clean         Remove local build artifacts"
	@echo ""
	@echo "Schema Management (Atlas):"
	@echo "  make schema-diff            Preview schema changes (production DB vs schema.sql)"
	@echo "  make schema-apply           Apply schema.sql to production database"
	@echo "  make schema-inspect         Dump current production database schema"
	@echo "  make schema-diff-workbench  Preview schema changes (workbench DB vs schema.sql)"
	@echo "  make schema-apply-workbench Apply schema.sql to workbench database"
	@echo "  make setup-workbench        Create/reset workbench-local database"
	@echo ""
	@echo "Installation:"
	@echo "  make install    Install global orc-bin + local-first shim"
	@echo "  make uninstall-shim  Remove shim, restore direct binary"
	@echo ""
	@echo "The shim prefers ./orc when present, falls back to orc-bin."
	@echo ""
	@echo "Claude Code Integration:"
	@echo "  make deploy-glue   Deploy skills to ~/.claude/skills/"
