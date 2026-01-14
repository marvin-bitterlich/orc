#!/bin/bash
# cleanup-test-env.sh - Clean up test environment
# Usage: ./cleanup-test-env.sh <mission-id>

set -euo pipefail

MISSION_ID="${1:-}"

if [[ -z "$MISSION_ID" ]]; then
    echo "status=ERROR"
    echo "error=Missing mission ID argument"
    exit 1
fi

echo "=== Cleanup Starting ===" >&2
echo "mission_id=$MISSION_ID"

MISSION_DIR="$HOME/src/missions/$MISSION_ID"
CLEANUP_SUCCESS=true

# 1. Kill TMux session if exists
echo "=== TMux Session Cleanup ===" >&2

SESSION_NAME="orc-$MISSION_ID"

if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
    if tmux kill-session -t "$SESSION_NAME" 2>/dev/null; then
        echo "tmux_session_killed=true"
    else
        echo "tmux_session_killed=false"
        echo "warning=Failed to kill TMux session"
        CLEANUP_SUCCESS=false
    fi
else
    echo "tmux_session_killed=not_found"
    echo "info=TMux session did not exist"
fi

# 2. Find and remove grove(s) for this mission
echo "=== Grove Cleanup ===" >&2

# List groves for this mission
if cd "$MISSION_DIR" 2>/dev/null; then
    GROVE_LIST=$(orc grove list 2>/dev/null || echo "")

    if [[ -n "$GROVE_LIST" ]]; then
        # Extract grove IDs and paths
        GROVE_COUNT=0

        while IFS= read -r line; do
            if echo "$line" | grep -q "test-canary"; then
                GROVE_ID=$(echo "$line" | awk '{print $1}')
                GROVE_PATH=$(echo "$line" | awk '{print $NF}')

                ((GROVE_COUNT++))

                # Remove git worktree
                if [[ -d "$GROVE_PATH" ]]; then
                    REPO_PATH=$(dirname "$(dirname "$GROVE_PATH")")/orc-canary

                    if [[ -d "$REPO_PATH" ]]; then
                        cd "$REPO_PATH"

                        if git worktree remove "$GROVE_PATH" --force 2>/dev/null; then
                            echo "grove_${GROVE_COUNT}_worktree_removed=true"
                        else
                            echo "grove_${GROVE_COUNT}_worktree_removed=false"
                            echo "warning=Failed to remove worktree: $GROVE_PATH"

                            # Try manual cleanup
                            rm -rf "$GROVE_PATH" 2>/dev/null || true
                        fi

                        # Prune worktrees
                        git worktree prune 2>/dev/null || true
                    fi
                fi

                # TODO: Remove from ORC database when orc grove delete command exists
                # For now, just note it
                echo "grove_${GROVE_COUNT}_id=$GROVE_ID"
                echo "grove_${GROVE_COUNT}_db_cleaned=pending"
                echo "info=Grove $GROVE_ID not removed from database (orc grove delete not implemented)"
            fi
        done <<< "$GROVE_LIST"

        echo "groves_found=$GROVE_COUNT"
    else
        echo "groves_found=0"
        echo "info=No groves found for mission"
    fi
else
    echo "groves_found=unknown"
    echo "warning=Could not access mission directory to list groves"
fi

# 3. Remove mission directory
echo "=== Mission Directory Cleanup ===" >&2

if [[ -d "$MISSION_DIR" ]]; then
    if rm -rf "$MISSION_DIR" 2>/dev/null; then
        echo "mission_dir_removed=true"
    else
        echo "mission_dir_removed=false"
        echo "error=Failed to remove mission directory: $MISSION_DIR"
        CLEANUP_SUCCESS=false
    fi
else
    echo "mission_dir_removed=not_found"
    echo "info=Mission directory did not exist"
fi

# 4. Remove mission from database
echo "=== Mission Database Cleanup ===" >&2

# TODO: Call orc mission delete when implemented
# For now, just note it
echo "mission_db_cleaned=pending"
echo "info=Mission $MISSION_ID not removed from database (orc mission delete not implemented)"

# Overall status
echo "=== Cleanup Complete ===" >&2

if $CLEANUP_SUCCESS; then
    echo "status=OK"
else
    echo "status=PARTIAL"
    echo "warning=Some cleanup steps failed"
fi

echo "timestamp=$(date -u +\"%Y-%m-%dT%H:%M:%SZ\")"

exit 0
