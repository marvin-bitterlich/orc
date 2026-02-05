# ORC Journal

Capture observations, friction, and learnings as journal notes at commission level.

## Usage

```
/journal "My observation about X"
```

## Behavior

1. Get focused entity from `orc focus --show`
2. Extract commission ID (directly if commission focused, or from shipment's commission)
3. If no focus: error "Focus a commission or shipment first"
4. Create note: `orc note create "<observation>" --commission <COMM-xxx> --type journal`
5. Output: "Journal entry created: NOTE-xxx"

## Flow

```bash
# Get current focus
orc focus --show
```

Parse output to extract commission ID:
- If focused on commission: use that ID
- If focused on shipment: get commission from shipment
- If no focus: error

```bash
# Create journal note at commission level
orc note create "<title>" --commission <COMM-xxx> --type journal
```

## Example

```
> /journal "CLI feels clunky when moving between shipments"

Journal entry created: NOTE-602
  Type: journal
  Commission: COMM-001
```

## Notes

- Journal notes are always created at commission level (not shipment)
- Type is always 'journal'
- Minimal friction - just capture and continue
- Good for friction, observations, ideas during work
