---
name: ship-verify
description: |
  Post-deploy verification for shipments. Use when user says "/ship-verify",
  "verify deployment", "verify shipment", or wants to confirm a deploy was successful.
  Runs verification checks and transitions shipment to verified status.
---

# Ship Verify

Post-deploy verification that runs smoke tests and transitions shipment to verified status.

## Workflow

### 1. Get Deployed Shipment

```bash
orc status  # Get focused shipment
orc shipment show SHIP-XXX
```

Verify shipment status is `deployed`. If not deployed:
- If `implemented`: "Shipment not yet deployed. Run /ship-deploy first."
- If `verified`: "Shipment already verified."
- If `complete`: "Shipment already complete."
- Other: "Shipment must be in deployed status to verify."

### 2. Detect Build System

Infer test commands from build system:

| Build System | Test Command |
|--------------|--------------|
| Makefile | `make test` |
| package.json | `npm test` |
| Gemfile | `bundle exec rspec` |

If no build system found:
```
Warning: No build system detected. Skipping automated tests.
```

### 3. Run Verification Checks

Run available verification commands:

```bash
# ORC status checks (always run if orc available)
orc status
orc commission list
orc shipment list

# Build system tests (if detected)
<test-command>  # e.g., make test, npm test, bundle exec rspec
```

Report each check as PASS/FAIL:
```
Verification Results:
  [PASS] orc status
  [PASS] orc commission list
  [PASS] orc shipment list
  [PASS] <test-command>
```

Or if no build system:
```
Verification Results:
  [PASS] orc status
  [PASS] orc commission list
  [PASS] orc shipment list
  [SKIP] tests - no build system detected
```

### 4. Handle Results

**If all checks pass:**
```bash
orc shipment verify SHIP-XXX
```

Output:
```
Shipment SHIP-XXX verified successfully.

Next steps:
  /ship-complete SHIP-XXX  # Complete shipment (terminal state)
```

**If any check fails:**
Do NOT transition status. Report:
```
Verification failed:
  [FAIL] <test-command> - exit code 1

Fix the failing checks and re-run /ship-verify
```

### 5. Notify Goblin (Optional)

If verification passes and in a workshop context, optionally notify goblin:

```bash
orc workshop show  # Get gatehouse ID
orc mail send "SHIP-XXX verified - all checks pass" \
  --to GOBLIN-GATE-XXX \
  --subject "SHIP-XXX Verified"
```

## Success Output

```
Verifying SHIP-XXX: Feature Name

Detecting build system...
  Found: Makefile

Running verification checks...
  [PASS] orc status
  [PASS] orc commission list
  [PASS] orc shipment list
  [PASS] make test

All checks passed!

Transitioning shipment to verified...
✓ Shipment SHIP-XXX verified

Next steps:
  /ship-complete SHIP-XXX  # Complete shipment (terminal state)
```

## Error Handling

| Error | Action |
|-------|--------|
| Shipment not deployed | Report required status, suggest /ship-deploy |
| Check fails | Report which check failed, do not transition |
| No focused shipment | Ask for shipment ID or suggest `orc focus SHIP-XXX` |
| No build system | Warn and skip test step, continue with other checks |

## Usage

```
/ship-verify              (verify focused shipment)
/ship-verify SHIP-xxx     (verify specific shipment)
```
