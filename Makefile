.PHONY: install install-binary install-shim uninstall-shim dev build test clean help

# Go binary location (handles empty GOBIN)
GOBIN := $(shell go env GOPATH)/bin

# Default target
.DEFAULT_GOAL := help

#---------------------------------------------------------------------------
# Installation (global binary + shim)
#---------------------------------------------------------------------------

# Full install: binary + shim
install: install-binary install-shim
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
	go build -o $(GOBIN)/orc-bin ./cmd/orc

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
	@go build -o orc ./cmd/orc
	@echo "✓ Built ./orc (local development binary)"

# Alias for backwards compatibility
build: dev

#---------------------------------------------------------------------------
# Testing & Maintenance
#---------------------------------------------------------------------------

# Run all tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f orc
	go clean
	@echo "✓ Cleaned local build artifacts"

#---------------------------------------------------------------------------
# Help
#---------------------------------------------------------------------------

help:
	@echo "ORC Makefile Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev        Build local ./orc for development"
	@echo "  make test       Run all tests"
	@echo "  make clean      Remove local build artifacts"
	@echo ""
	@echo "Installation:"
	@echo "  make install    Install global orc-bin + local-first shim"
	@echo "  make uninstall-shim  Remove shim, restore direct binary"
	@echo ""
	@echo "The shim prefers ./orc when present, falls back to orc-bin."
