#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
BASE_IMAGE="ghcr.io/cirruslabs/macos-tahoe-base:latest"
VM_NAME="orc-bootstrap-test-$$"
VM_USER="admin"
VM_PASS="admin"
SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR"

# Flags
KEEP=false
KEEP_ON_FAILURE=false
SHELL_MODE=false
VERBOSE=false

# Timing
START_TIME=$(date +%s)

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Test 'make bootstrap' in a clean macOS Tart VM.

Options:
    --shell             Drop into interactive shell after bootstrap (implies --keep)
    --keep              Keep VM after test (success or failure)
    --keep-on-failure   Keep VM only on failure for debugging
    --verbose, -v       Show verbose output
    --help, -h          Show this help message

Requirements:
    - tart (https://github.com/cirruslabs/tart)
    - ssh

Examples:
    $(basename "$0")                    # Run bootstrap test
    $(basename "$0") --shell            # Bootstrap then drop into VM shell
    $(basename "$0") --keep             # Keep VM for exploration
    $(basename "$0") --keep-on-failure  # Keep VM if test fails
EOF
}

log() {
    local elapsed=$(($(date +%s) - START_TIME))
    printf "[%3ds] %s\n" "$elapsed" "$1"
}

log_verbose() {
    if [[ "$VERBOSE" == "true" ]]; then
        log "$1"
    fi
}

error() {
    log "✗ ERROR: $1" >&2
}

print_vm_info() {
    log ""
    log "VM kept for exploration:"
    log "  VM Name:  $VM_NAME"
    log "  SSH:      sshpass -p $VM_PASS ssh $VM_USER@$VM_IP"
    log "  Password: $VM_PASS"
    log ""
    log "  Cleanup when done:"
    log "    tart stop $VM_NAME"
    log "    tart delete $VM_NAME"
}

cleanup() {
    local exit_code=$?

    # Always keep if --keep was specified
    if [[ "$KEEP" == "true" ]]; then
        return
    fi

    # Keep on failure if --keep-on-failure was specified
    if [[ "$exit_code" -ne 0 ]] && [[ "$KEEP_ON_FAILURE" == "true" ]]; then
        log "⚠ Keeping VM '$VM_NAME' for debugging (--keep-on-failure)"
        log "  To SSH: ssh $VM_USER@\$(tart ip $VM_NAME)"
        log "  To delete: tart stop $VM_NAME && tart delete $VM_NAME"
        return
    fi

    if tart list 2>/dev/null | grep -q "^$VM_NAME"; then
        log "Cleaning up VM '$VM_NAME'..."
        tart stop "$VM_NAME" 2>/dev/null || true
        tart delete "$VM_NAME" 2>/dev/null || true
    fi
}

trap cleanup EXIT

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --shell)
            SHELL_MODE=true
            KEEP=true  # --shell implies --keep
            shift
            ;;
        --keep)
            KEEP=true
            shift
            ;;
        --keep-on-failure)
            KEEP_ON_FAILURE=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Check dependencies
log "Checking dependencies..."

if ! command -v tart &>/dev/null; then
    error "tart not found"
    echo ""
    echo "Install tart with:"
    echo "  brew install cirruslabs/cli/tart"
    echo ""
    echo "More info: https://github.com/cirruslabs/tart"
    exit 1
fi

if ! command -v ssh &>/dev/null; then
    error "ssh not found"
    exit 1
fi

if ! command -v sshpass &>/dev/null; then
    error "sshpass not found"
    echo ""
    echo "Install sshpass with:"
    echo "  brew install sshpass"
    exit 1
fi

log "✓ Dependencies OK"

# Clean up any orphan test VMs from previous runs
ORPHANS=$(tart list 2>/dev/null | grep "orc-bootstrap-test-" | awk '{print $2}' || true)
if [[ -n "$ORPHANS" ]]; then
    log "Cleaning up orphan test VMs..."
    for vm in $ORPHANS; do
        log_verbose "  Stopping $vm"
        tart stop "$vm" 2>/dev/null || true
        tart delete "$vm" 2>/dev/null || true
    done
    log "✓ Orphans cleaned up"
fi

# Pull base image if needed
if ! tart list 2>/dev/null | grep -q "tahoe-base"; then
    log "Pulling base image (this may take a while)..."
    tart pull "$BASE_IMAGE"
    log "✓ Base image pulled"
else
    log "✓ Base image already present"
fi

# Clone VM
log "Creating test VM '$VM_NAME'..."
tart clone ghcr.io/cirruslabs/macos-tahoe-base:latest "$VM_NAME"
log "✓ VM created"

# Start VM headless with shared directory
log "Starting VM headless with shared ORC repo..."
tart run "$VM_NAME" --no-graphics --dir="orc:$PROJECT_ROOT" &
VM_PID=$!

# Wait for VM to boot and get IP
log "Waiting for VM to boot..."
VM_IP=""
for i in {1..60}; do
    VM_IP=$(tart ip "$VM_NAME" 2>/dev/null || true)
    if [[ -n "$VM_IP" ]]; then
        break
    fi
    sleep 2
    log_verbose "  Waiting for IP... ($i/60)"
done

if [[ -z "$VM_IP" ]]; then
    error "Failed to get VM IP after 120 seconds"
    exit 1
fi

log "✓ VM booted (IP: $VM_IP)"

# Wait for SSH to be ready
log "Waiting for SSH..."
for i in {1..30}; do
    if sshpass -p "$VM_PASS" ssh $SSH_OPTS "$VM_USER@$VM_IP" "echo ready" &>/dev/null; then
        break
    fi
    sleep 2
    log_verbose "  Waiting for SSH... ($i/30)"
done

if ! sshpass -p "$VM_PASS" ssh $SSH_OPTS "$VM_USER@$VM_IP" "echo ready" &>/dev/null; then
    error "SSH not available after 60 seconds"
    exit 1
fi

log "✓ SSH ready"

# Run bootstrap sequence via SSH
# Use login shell to ensure PATH includes Homebrew
run_ssh() {
    sshpass -p "$VM_PASS" ssh $SSH_OPTS "$VM_USER@$VM_IP" "zsh -l -c '$*'"
}

log "Installing Go via Homebrew..."
run_ssh "brew install go" || {
    error "Failed to install Go"
    exit 1
}
log "✓ Go installed"

log "Copying ORC repo to ~/src/orc (canonical location)..."
# Copy repo to canonical location ~/src/orc
# Exclude .git (worktrees have references that won't work in VM)
# Then init fresh git repo for make bootstrap
run_ssh "mkdir -p ~/src/orc && rsync -a --exclude .git /Volumes/My\ Shared\ Files/orc/ ~/src/orc/ && cd ~/src/orc && git init && git add -A && git commit -m 'Initial commit for bootstrap test'" || {
    error "Failed to copy repo"
    exit 1
}
log "✓ Repo copied to ~/src/orc"

log "Creating Claude settings.json stub..."
sshpass -p "$VM_PASS" ssh $SSH_OPTS "$VM_USER@$VM_IP" 'mkdir -p ~/.claude && echo '"'"'{"hooks": {}}'"'"' > ~/.claude/settings.json'
log "✓ Claude settings.json created"

log "Running 'make bootstrap'..."
if run_ssh "cd ~/src/orc && make bootstrap"; then
    log "✓ make bootstrap PASSED"
else
    error "make bootstrap FAILED"
    exit 1
fi

# Verify orc is in PATH (fresh login shell sources updated ~/.zshrc)
log "Verifying orc is in PATH..."
if run_ssh "orc --version"; then
    log "✓ orc command works via PATH"
else
    error "orc not found in PATH after bootstrap"
    exit 1
fi

# Verify FACT-001 was created by make bootstrap
log "Verifying FACT-001 exists..."
if run_ssh "orc factory list | grep -q FACT-001"; then
    log "✓ FACT-001 exists"
else
    error "FACT-001 not found after bootstrap"
    exit 1
fi

# Verify REPO-001 was created by make bootstrap
log "Verifying REPO-001 exists..."
if run_ssh "orc repo list | grep -q REPO-001"; then
    log "✓ REPO-001 exists"
else
    error "REPO-001 not found after bootstrap"
    exit 1
fi

# Verify REPO-001 has correct path
log "Verifying REPO-001 path..."
if run_ssh "orc repo show REPO-001 | grep -q 'src/orc'"; then
    log "✓ REPO-001 has correct path"
else
    error "REPO-001 path incorrect (expected ~/src/orc)"
    exit 1
fi

# Verify CLI functionality with real commands
log "Verifying CLI functionality..."

log "Creating test commission..."
if run_ssh "orc commission create 'Bootstrap Test'"; then
    log "✓ Commission created"
else
    error "Failed to create commission"
    exit 1
fi

log "Creating test workshop..."
if run_ssh "orc workshop create 'Test Workshop' --factory FACT-001"; then
    log "✓ Workshop created"
else
    error "Failed to create workshop"
    exit 1
fi

log "Running orc summary..."
if run_ssh "orc summary"; then
    log "✓ Summary works"
else
    error "Failed to run summary"
    exit 1
fi

log "Running orc doctor..."
if run_ssh "orc doctor"; then
    log "✓ Doctor passes"
else
    error "orc doctor reported issues"
    exit 1
fi

# Final timing
ELAPSED=$(($(date +%s) - START_TIME))
log ""
log "=========================================="
log "✓ Bootstrap test PASSED in ${ELAPSED}s"
log "=========================================="

# Print VM info if keeping
if [[ "$KEEP" == "true" ]]; then
    print_vm_info
fi

# Drop into shell if requested
if [[ "$SHELL_MODE" == "true" ]]; then
    log ""
    log "Dropping into VM shell..."
    exec sshpass -p "$VM_PASS" ssh $SSH_OPTS "$VM_USER@$VM_IP"
fi

exit 0
