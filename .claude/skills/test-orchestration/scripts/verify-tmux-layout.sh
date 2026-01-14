#!/bin/bash
# verify-tmux-layout.sh - Verify TMux session structure
# Usage: ./verify-tmux-layout.sh <session-name>

set -euo pipefail

SESSION_NAME="${1:-}"

if [[ -z "$SESSION_NAME" ]]; then
    echo "status=ERROR"
    echo "error=Missing session name argument"
    exit 1
fi

# Check if session exists
if ! tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
    echo "status=FAIL"
    echo "session_exists=false"
    echo "error=Session '$SESSION_NAME' does not exist"
    exit 1
fi

echo "session_exists=true"

# Count windows
WINDOW_COUNT=$(tmux list-windows -t "$SESSION_NAME" 2>/dev/null | wc -l | tr -d ' ')
echo "window_count=$WINDOW_COUNT"

# Check for deputy window
if tmux list-windows -t "$SESSION_NAME" 2>/dev/null | grep -q "deputy"; then
    echo "deputy_window=true"
else
    echo "deputy_window=false"
fi

# Check for IMP window
if tmux list-windows -t "$SESSION_NAME" 2>/dev/null | grep -q "imp"; then
    echo "imp_window=true"

    # Count panes in IMP window
    IMP_PANE_COUNT=$(tmux list-panes -t "$SESSION_NAME:imp" 2>/dev/null | wc -l | tr -d ' ')
    echo "imp_pane_count=$IMP_PANE_COUNT"

    # Expected IMP layout is 3 panes
    if [[ "$IMP_PANE_COUNT" == "3" ]]; then
        echo "imp_layout=valid"
    else
        echo "imp_layout=invalid"
    fi
else
    echo "imp_window=false"
    echo "imp_pane_count=0"
    echo "imp_layout=missing"
fi

# Total pane count across all windows
TOTAL_PANES=$(tmux list-panes -t "$SESSION_NAME" -a 2>/dev/null | wc -l | tr -d ' ')
echo "total_panes=$TOTAL_PANES"

# Determine overall status
if [[ "$WINDOW_COUNT" -ge 2 ]] && \
   tmux list-windows -t "$SESSION_NAME" 2>/dev/null | grep -q "deputy" && \
   tmux list-windows -t "$SESSION_NAME" 2>/dev/null | grep -q "imp"; then
    echo "status=OK"
    echo "layout=valid"
else
    echo "status=FAIL"
    echo "layout=invalid"
fi

exit 0
