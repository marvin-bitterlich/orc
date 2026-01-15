# AGENTS.md - Development Rules for Claude Agents

This file contains essential development workflow rules for Claude agents working on the ORC codebase.

## Build & Install

**ALWAYS use the Makefile for building and installing ORC:**

```bash
make install    # Build and install globally to $GOPATH/bin/orc
make build      # Build locally to ./orc
make test       # Run tests
make clean      # Clean build artifacts
```

**DO NOT manually run `go build` commands.** The Makefile handles correct paths and prevents installation errors.

**Why this matters:**
- Prevents repeated mistakes with `$GOPATH/bin/orc` vs `/usr/local/bin/orc` vs `~/bin/orc`
- Reduces cognitive load - one consistent command
- Avoids permission errors (no sudo needed for $GOPATH/bin)
- $GOPATH/bin is already in PATH

## Common Mistakes to Avoid

❌ `go build -o $GOPATH/bin/orc ./cmd/orc` ← Don't do this manually
✅ `make install` ← Use this instead

❌ `sudo cp orc /usr/local/bin/orc` ← Don't use sudo
✅ `make install` ← Installs to user's GOPATH

❌ Forgetting to rebuild after code changes
✅ `make install` after every change

## Quick Reference

After making code changes:
```bash
make install && orc summary --mission current
```

That's it. No paths to remember, no permissions to fight with.
