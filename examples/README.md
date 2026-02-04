# Example OpenCode Configuration Files

This directory contains example configuration files for Gas Town / OpenCode setup.

## Configuration Architecture

Gas Town uses a **layered configuration system** with multiple sources:

### 1. Role Definitions (TOML)
**Location:** `internal/config/roles/*.toml`

These define agent behavior, session patterns, and health settings. They are **code-level configuration** baked into the binary.

See: [roles/](./roles/)

### 2. Rig Identity (config.json)
**Location:** `<rig>/config.json`

Contains rig identity information (type, name, git_url, beads prefix). Created when you run `gt rig add`.

See: [rig-config.json](./rig-config.json)

### 3. Rig Settings (settings/config.json)
**Location:** `<rig>/settings/config.json`

Contains behavioral configuration like agent preferences, themes, and merge queue settings.

See: [rig-settings.json](./rig-settings.json)

### 4. Town Settings (~/gt/settings/config.json)
**Location:** `~/gt/settings/config.json`

Town-wide defaults that apply to all rigs.

See: [town-settings.json](./town-settings.json)

### 5. Runtime Configuration (Wisp Layer)
**Location:** `<rig>/.beads-wisp/config/`

Transient, local-only overrides for temporary adjustments.

See: [wisp-config.json](./wisp-config.json)

### 6. Rig Operational State (Bead Labels)
**Location:** Stored as labels on the rig identity bead in `<rig>/.beads/`

Persistent, synced operational state like `status:docked` or `priority_adjustment:10`.

Managed via: `gt rig config set <rig> <key> <value> --global`

## Quick Start

1. **Set up a new rig:**
   ```bash
   gt rig add myproject https://github.com/user/myproject.git
   ```

2. **Configure rig settings:**
   ```bash
   # Copy example and customize
   cp examples/rig-settings.json myrig/settings/config.json
   
   # Or use CLI
   gt rig settings set myproject agent opencode
   ```

3. **Set operational state:**
   ```bash
   # Temporary (local only)
   gt rig config set myproject status parked
   
   # Permanent (synced to all clones)
   gt rig config set myproject max_polecats 8 --global
   ```

## Important Distinction

| Config Type | Purpose | Mutable | Synced |
|-------------|---------|---------|--------|
| Role TOML | Agent definitions | No (code) | N/A |
| config.json | Rig identity | Rarely | Yes (git) |
| settings/config.json | Behavioral prefs | Yes | Yes (git) |
| Wisp layer | Local overrides | Yes | No |
| Bead labels | Operational state | Yes | Yes (beads sync) |

## Documentation Discrepancy Note

**IMPORTANT:** The documentation in `docs/reference.md` and `docs/design/property-layers.md` contains outdated information:

- Docs mention `config.json` for rig config - **partially correct**, it's for identity only
- Docs don't mention **role TOML files** at all - these are the actual source of agent configuration
- The `gt rig config` commands work with **bead labels** and **wisp layer**, not JSON files

See [CONFIG_DISCREPANCY.md](./CONFIG_DISCREPANCY.md) for full details.
