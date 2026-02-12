# Troubleshooting

**Status**: Living document
**Last Updated**: 2026-02-09

Two entry points when you're stuck.

---

## Skill Discovery

Don't know which skill to use?

```
/orc-help
```

Shows categorized skill list with examples:
- `/ship-*` — Shipment lifecycle
- `/imp-*` — Implementation workflow
- `/orc-*` — Utilities and setup

---

## Environment Health

Something broken?

```bash
orc doctor
```

Checks:
- Database connection
- Git configuration
- Claude Code integration
- Skills deployment

Fix most issues by running what `orc doctor` suggests.
