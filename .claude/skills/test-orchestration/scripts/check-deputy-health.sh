#!/bin/bash
# check-deputy-health.sh - Verify deputy ORC is operational
# Usage: ./check-deputy-health.sh <mission-id>

set -euo pipefail

MISSION_ID="${1:-}"

if [[ -z "$MISSION_ID" ]]; then
    echo "status=ERROR"
    echo "error=Missing mission ID argument"
    exit 1
fi

MISSION_DIR="$HOME/src/missions/$MISSION_ID"

# Check if mission directory exists
if [[ ! -d "$MISSION_DIR" ]]; then
    echo "status=FAIL"
    echo "mission_dir_exists=false"
    echo "error=Mission directory not found: $MISSION_DIR"
    exit 1
fi

echo "mission_dir_exists=true"

# Check for .orc-mission marker
if [[ ! -f "$MISSION_DIR/.orc-mission" ]]; then
    echo "status=FAIL"
    echo "mission_marker_exists=false"
    echo "error=.orc-mission marker not found"
    exit 1
fi

echo "mission_marker_exists=true"

# Validate .orc-mission JSON
if ! jq -e . "$MISSION_DIR/.orc-mission" >/dev/null 2>&1; then
    echo "status=FAIL"
    echo "mission_marker_valid=false"
    echo "error=.orc-mission is not valid JSON"
    exit 1
fi

echo "mission_marker_valid=true"

# Extract mission ID from marker
MARKER_MISSION_ID=$(jq -r '.mission_id' "$MISSION_DIR/.orc-mission")
if [[ "$MARKER_MISSION_ID" != "$MISSION_ID" ]]; then
    echo "status=FAIL"
    echo "mission_id_match=false"
    echo "error=Mission ID mismatch: expected $MISSION_ID, got $MARKER_MISSION_ID"
    exit 1
fi

echo "mission_id_match=true"

# Check workspace metadata
if [[ ! -f "$MISSION_DIR/.orc/metadata.json" ]]; then
    echo "status=WARN"
    echo "metadata_exists=false"
    echo "warning=metadata.json not found (not critical)"
else
    echo "metadata_exists=true"

    # Validate metadata JSON
    if ! jq -e . "$MISSION_DIR/.orc/metadata.json" >/dev/null 2>&1; then
        echo "status=WARN"
        echo "metadata_valid=false"
        echo "warning=metadata.json is not valid JSON"
    else
        echo "metadata_valid=true"

        # Check active_mission_id matches
        ACTIVE_MISSION=$(jq -r '.active_mission_id' "$MISSION_DIR/.orc/metadata.json")
        if [[ "$ACTIVE_MISSION" == "$MISSION_ID" ]]; then
            echo "active_mission_match=true"
        else
            echo "active_mission_match=false"
            echo "warning=active_mission_id mismatch"
        fi
    fi
fi

# Test orc status from mission directory
cd "$MISSION_DIR"
if orc status 2>&1 | grep -q "$MISSION_ID"; then
    echo "status_command=working"
    echo "status_shows_mission=true"
else
    echo "status_command=working"
    echo "status_shows_mission=false"
    echo "warning=orc status does not show mission ID"
fi

# Check if deputy context detected
if orc status 2>&1 | grep -qi "deputy"; then
    echo "context=deputy"
    echo "context_detected=true"
else
    echo "context=unknown"
    echo "context_detected=false"
    echo "warning=Deputy context not detected"
fi

# Test orc summary
if orc summary >/dev/null 2>&1; then
    echo "summary_command=working"
else
    echo "summary_command=failed"
    echo "warning=orc summary command failed"
fi

# Overall status
if [[ -f "$MISSION_DIR/.orc-mission" ]] && \
   jq -e . "$MISSION_DIR/.orc-mission" >/dev/null 2>&1 && \
   orc status 2>&1 | grep -q "$MISSION_ID"; then
    echo "status=OK"
    echo "mission=$MISSION_ID"
else
    echo "status=FAIL"
    echo "mission=$MISSION_ID"
fi

exit 0
