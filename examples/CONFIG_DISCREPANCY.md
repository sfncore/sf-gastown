# Configuration Documentation Discrepancy

## Summary

The Gas Town documentation contains significant inaccuracies regarding the configuration system. This document details the discrepancies between documented behavior and actual implementation.

## Documented vs Actual

### What the Docs Say

The documentation in `docs/reference.md` and `docs/design/property-layers.md` describes:

1. **Rig Config (`config.json`)** - Described as containing rig configuration
2. **`gt rig config` commands** - Described as manipulating JSON configuration
3. **Four-layer property system** - Wisp, Bead, Town, System layers

### What Actually Exists

1. **Role TOML Files** (`internal/config/roles/*.toml`)
   - Define agent roles (mayor, witness, polecat, refinery, deacon, crew, dog)
   - Contain session patterns, environment variables, health settings
   - **NOT MENTIONED IN DOCS**

2. **Rig Identity (`config.json`)**
   - Contains: type, name, git_url, beads prefix
   - Created by `gt rig add`
   - Docs are partially correct but don't clarify limited scope

3. **Rig Settings (`settings/config.json`)**
   - Contains: agent preferences, themes, merge_queue settings
   - Docs conflate this with `gt rig config` commands

4. **`gt rig config` Commands**
   - Work with **bead labels** on rig identity beads
   - Work with **wisp layer** (`.beads-wisp/config/`)
   - Do NOT modify JSON files directly

## Specific Discrepancies

### 1. Role Configuration

**Docs:** No mention of role configuration files

**Reality:** Seven TOML files in `internal/config/roles/`:
- `mayor.toml` - Town coordinator
- `witness.toml` - Worker monitor
- `polecat.toml` - Ephemeral workers
- `refinery.toml` - Merge queue processor
- `deacon.toml` - Daemon beacon
- `crew.toml` - Persistent workspaces
- `dog.toml` - Cross-rig workers

Each contains:
```toml
role = "polecat"
scope = "rig"
nudge = "Check your hook for work assignments."
prompt_template = "polecat.md.tmpl"

[session]
pattern = "gt-{rig}-{name}"
work_dir = "{town}/{rig}/polecats/{name}"
needs_pre_sync = true
start_command = "exec claude --dangerously-skip-permissions"

[env]
GT_ROLE = "polecat"
GT_SCOPE = "rig"

[health]
ping_timeout = "30s"
consecutive_failures = 3
kill_cooldown = "5m"
stuck_threshold = "2h"
```

### 2. `gt rig config` Commands

**Docs (property-layers.md):**
```bash
gt rig config show gastown           # Show effective config
gt rig config set gastown key value  # Set in wisp layer
gt rig config set gastown key value --global  # Set in bead layer
```

**Reality:**
- `gt rig config show` - Shows merged config from wisp + bead labels + defaults
- `gt rig config set` - Sets values in wisp layer (`.beads-wisp/config/<rig>.json`)
- `gt rig config set --global` - Sets labels on rig identity bead (stored in beads DB)

The "bead layer" is **NOT** a JSON file - it's labels on the rig identity bead in the beads database.

### 3. Configuration Keys

**Docs mention:**
- `status` (operational/parked/docked)
- `auto_restart`
- `max_polecats`
- `priority_adjustment`
- `maintenance_window`
- `dnd`

**Reality:**
These are operational properties stored via bead labels, NOT JSON configuration. The actual JSON config files contain different keys.

### 4. Config File Locations

**Docs (reference.md):**
```
<rig>/
├── config.json             # Rig configuration
└── settings/
    └── config.json         # Settings
```

**Reality:**
- `<rig>/config.json` - Rig identity (name, git_url, beads prefix)
- `<rig>/settings/config.json` - Behavioral settings (agents, theme, merge_queue)
- `<rig>/.beads-wisp/config/<rig>.json` - Wisp layer (transient)
- `<rig>/.beads/` - Beads database with rig identity bead labels

## The Four Layers (Corrected)

| Layer | Location | Actual Implementation |
|-------|----------|----------------------|
| 1. Wisp | `.beads-wisp/config/<rig>.json` | JSON file, local only |
| 2. Bead | `.beads/` (rig identity bead labels) | Beads DB labels, synced via `bd sync` |
| 3. Town | `~/gt/settings/config.json` | JSON file, town-wide |
| 4. System | Compiled-in | Go code in `internal/rig/config.go` |

**Missing from docs:** Role TOML files are the PRIMARY configuration for agent behavior.

## Impact

1. **Users following docs** will look for JSON files that don't control what they expect
2. **Role customization** requires code changes (TOML files are compiled in)
3. **Operational state** (parked/docked) is correctly documented but confused with static config

## Recommendations

1. Update `docs/reference.md` to document role TOML files
2. Clarify distinction between:
   - Role configuration (TOML, code-level)
   - Rig identity (config.json)
   - Rig settings (settings/config.json)
   - Operational state (bead labels, wisp layer)
3. Update `docs/design/property-layers.md` to remove references to JSON manipulation via `gt rig config`
4. Document that `gt rig config` works with bead labels, not JSON files
