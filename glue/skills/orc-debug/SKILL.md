---
name: orc-debug
description: View recent Claude Code tool calls from the ORC debug log. Use when you want to see what tools have been called, debug hook behavior, or inspect tool usage across sessions.
---

# ORC Debug Log

Show recent tool calls from the debug log.

## Usage

`/orc-debug` - show last 30 entries
`/orc-debug 50` - show last 50 entries

## Action

If user provides a number, use that as the tail count. Otherwise default to 30.

Run:
```bash
tail -N ~/.claude/orc-debug.log
```

Where N is the count (default 30).

Display the output to the user.
