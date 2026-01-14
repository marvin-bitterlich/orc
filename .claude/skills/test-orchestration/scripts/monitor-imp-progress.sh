#!/bin/bash
# monitor-imp-progress.sh - Watch IMP activity and work order progress
# Usage: ./monitor-imp-progress.sh <session-name> [mission-id]

set -euo pipefail

SESSION_NAME="${1:-}"
MISSION_ID="${2:-}"

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

# Get work order counts if mission ID provided
if [[ -n "$MISSION_ID" ]]; then
    MISSION_DIR="$HOME/src/missions/$MISSION_ID"

    if [[ -d "$MISSION_DIR" ]]; then
        cd "$MISSION_DIR"

        # Count work orders by status
        WO_LIST=$(orc work-order list 2>/dev/null || echo "")

        if [[ -n "$WO_LIST" ]]; then
            TOTAL_WO=$(echo "$WO_LIST" | grep -c "WO-" || echo "0")
            READY=$(echo "$WO_LIST" | grep -c "\[ready\]" || echo "0")
            IN_PROGRESS=$(echo "$WO_LIST" | grep -c "\[in_progress\]" || echo "0")
            COMPLETED=$(echo "$WO_LIST" | grep -c "\[completed\]" || echo "0")
            BLOCKED=$(echo "$WO_LIST" | grep -c "\[blocked\]" || echo "0")

            echo "work_orders=$TOTAL_WO"
            echo "ready=$READY"
            echo "in_progress=$IN_PROGRESS"
            echo "completed=$COMPLETED"
            echo "blocked=$BLOCKED"

            # Calculate progress percentage
            if [[ "$TOTAL_WO" -gt 0 ]]; then
                PROGRESS=$((COMPLETED * 100 / TOTAL_WO))
                echo "progress_percent=$PROGRESS"
            else
                echo "progress_percent=0"
            fi
        else
            echo "work_orders=0"
            echo "ready=0"
            echo "in_progress=0"
            echo "completed=0"
            echo "blocked=0"
            echo "progress_percent=0"
        fi
    else
        echo "mission_dir_exists=false"
        echo "warning=Mission directory not found"
    fi
else
    echo "work_orders=unknown"
    echo "warning=Mission ID not provided, skipping work order check"
fi

# Check IMP window activity (capture last command/output)
if tmux list-windows -t "$SESSION_NAME" 2>/dev/null | grep -q "imp"; then
    echo "imp_window_exists=true"

    # Get content from IMP shell pane (pane 2, assuming vim|claude|shell layout)
    IMP_PANE="$SESSION_NAME:imp.2"

    # Check if pane exists
    if tmux list-panes -t "$SESSION_NAME:imp" 2>/dev/null | grep -q "^2:"; then
        echo "imp_shell_pane_exists=true"

        # Capture last few lines from pane
        LAST_LINES=$(tmux capture-pane -t "$IMP_PANE" -p -S -10 2>/dev/null | tail -5 || echo "")

        # Check for activity indicators
        if echo "$LAST_LINES" | grep -qi "error\|fail"; then
            echo "imp_status=error_detected"
        elif echo "$LAST_LINES" | grep -qi "success\|pass\|complete"; then
            echo "imp_status=success_detected"
        elif echo "$LAST_LINES" | grep -q "\\$"; then
            echo "imp_status=idle"
        else
            echo "imp_status=active"
        fi
    else
        echo "imp_shell_pane_exists=false"
        echo "imp_status=unknown"
    fi
else
    echo "imp_window_exists=false"
    echo "imp_status=window_missing"
fi

# Overall status
echo "status=OK"
echo "timestamp=$(date -u +\"%Y-%m-%dT%H:%M:%SZ\")"

exit 0
