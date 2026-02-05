# OpenCode Plugin Configuration for Gas Town

## Problem Statement

When spawning new polecats with `gt sling`, the OpenCode plugin is not being copied to the polecat's `.opencode/` directory. This causes agents to spawn without their `gt prime` context, leaving them idle and unaware of their role or assignments.

## How the Code Works

### Plugin Copy Flow

```
gt sling <bead> <rig>
  └─► SessionManager.Start()
        └─► config.ResolveRoleAgentConfig("polecat", townRoot, rigPath)
              └─► config.ResolveAgentConfig()
                    └─► config.lookupAgentConfig()
                          └─► config.RuntimeConfigFromPreset(AgentOpenCode)
                                └─► config.normalizeRuntimeConfig(rc)  ← Sets Hooks defaults
        └─► runtime.EnsureSettingsForRole(polecatHomeDir, "polecat", runtimeConfig)
              └─► Case: provider == "opencode"
                    └─► opencode.EnsurePluginAt(workDir, rc.Hooks.Dir, rc.Hooks.SettingsFile)
                          └─► Copy embedded plugin/gastown.js to .opencode/plugin/gastown.js
```

### Critical Code Path

**File: `internal/runtime/runtime.go`**

```go
func EnsureSettingsForRole(workDir, role string, rc *config.RuntimeConfig) error {
    if rc == nil {
        rc = config.DefaultRuntimeConfig()
    }

    // ⚠️ SILENT FAILURE POINT #1
    if rc.Hooks == nil {
        return nil  // Returns without error if Hooks not configured!
    }

    provider := rc.Hooks.Provider
    // ⚠️ SILENT FAILURE POINT #2
    if provider == "" || provider == "none" {
        return nil  // Returns without error if provider not set!
    }

    switch provider {
    case "claude":
        return claude.EnsureSettingsForRoleAt(...)
    case "opencode":
        // This is where the plugin copy happens
        return opencode.EnsurePluginAt(workDir, rc.Hooks.Dir, rc.Hooks.SettingsFile)
    }
    return nil
}
```

**File: `internal/opencode/plugin.go`**

```go
//go:embed plugin/gastown.js
var pluginFS embed.FS

func EnsurePluginAt(workDir, pluginDir, pluginFile string) error {
    // Skip if already exists
    pluginPath := filepath.Join(workDir, pluginDir, pluginFile)
    if _, err := os.Stat(pluginPath); err == nil {
        return nil  // Already exists, don't overwrite
    }

    // Read from embedded filesystem
    content, err := pluginFS.ReadFile("plugin/gastown.js")
    if err != nil {
        return fmt.Errorf("reading plugin template: %w", err)
    }

    // Write to destination
    if err := os.WriteFile(pluginPath, content, 0644); err != nil {
        return fmt.Errorf("writing plugin: %w", err)
    }
    return nil
}
```

### RuntimeConfig Hook Defaults

**File: `internal/config/types.go`**

The `normalizeRuntimeConfig()` function automatically sets hook defaults based on provider:

```go
func normalizeRuntimeConfig(rc *RuntimeConfig) *RuntimeConfig {
    // ... other defaults ...

    if rc.Hooks == nil {
        rc.Hooks = &RuntimeHooksConfig{}
    }

    if rc.Hooks.Provider == "" {
        rc.Hooks.Provider = defaultHooksProvider(rc.Provider)
        // "opencode" → "opencode"
        // "claude" → "claude"
    }

    if rc.Hooks.Dir == "" {
        rc.Hooks.Dir = defaultHooksDir(rc.Provider)
        // "opencode" → ".opencode/plugin"
        // "claude" → ".claude"
    }

    if rc.Hooks.SettingsFile == "" {
        rc.Hooks.SettingsFile = defaultHooksFile(rc.Provider)
        // "opencode" → "gastown.js"
        // "claude" → "settings.json"
    }
    // ...
}
```

## Required Settings Configuration

### Town-Level Settings

**File: `~/gt/settings/config.json`**

```json
{
  "type": "town-settings",
  "version": 1,
  "default_agent": "opencode",
  "runtime": {
    "provider": "opencode"
  }
}
```

**Fields:**
- `default_agent`: The default agent to use when not specified per-role
- `runtime.provider`: MUST be set to "opencode" to trigger hook configuration

### Rig-Level Settings

**File: `~/gt/<rig>/settings/config.json`**

```json
{
  "type": "rig-settings",
  "version": 1,
  "runtime": {
    "provider": "opencode"
  }
}
```

**Fields:**
- `runtime.provider": "opencode"` - Overrides town default for this rig

### Alternative: Per-Role Agent Assignment

**In town or rig settings:**

```json
{
  "type": "town-settings",
  "version": 1,
  "default_agent": "claude",
  "role_agents": {
    "polecat": "opencode",
    "witness": "opencode",
    "refinery": "claude",
    "mayor": "claude"
  }
}
```

This allows different roles to use different agents.

## Configuration Resolution Order

When `ResolveRoleAgentConfig(role, townRoot, rigPath)` is called:

1. **Rig's RoleAgents** - Check `rigSettings.RoleAgents[role]`
2. **Town's RoleAgents** - Check `townSettings.RoleAgents[role]`
3. **Rig's Agent** - Check `rigSettings.Agent`
4. **Town's DefaultAgent** - Check `townSettings.DefaultAgent`
5. **Fallback** - Use "claude"

Once the agent name is determined:
- Look up agent preset (built-in or custom)
- Create RuntimeConfig from preset
- Apply `normalizeRuntimeConfig()` to set defaults including Hooks

## Plugin File Location

The plugin file should be created at:

```
<rig>/polecats/<name>/.opencode/plugin/gastown.js
```

For example:
```
~/gt/sfgastown/polecats/jasper/.opencode/plugin/gastown.js
```

## Plugin Content

The plugin (`internal/opencode/plugin/gastown.js`) provides:

1. **Session Creation Hook**: Runs `gt prime` and `gt mail check --inject` on startup
2. **Session Compaction Hook**: Re-injects context after compaction
3. **Session Deletion Hook**: Records session costs
4. **Context Injection**: Injects Gas Town role and directory info into system prompt

Key functionality:
```javascript
const injectContext = async () => {
  await run("gt prime");  // ← This is what gives agents their context!
  if (autonomousRoles.has(role)) {
    await run("gt mail check --inject");
  }
  await run("gt nudge deacon session-started");
};
```

## Debugging Checklist

If polecats are spawning without context:

1. **Check settings files exist:**
   ```bash
   cat ~/gt/settings/config.json
   cat ~/gt/<rig>/settings/config.json
   ```

2. **Verify runtime.provider is set:**
   ```bash
   grep -A2 '"runtime"' ~/gt/settings/config.json
   ```

3. **Check if plugin was copied:**
   ```bash
   ls -la ~/gt/<rig>/polecats/<name>/.opencode/plugin/
   ```

4. **Verify binary has embedded plugin:**
   ```bash
   strings $(which gt) | grep "gastown.js"
   ```

5. **Check for errors in polecat creation:**
   ```bash
   gt polecat status <rig>/<name>
   ```

## Common Issues

### Issue: "Database is locked by another dolt process"

**Cause**: Mixing embedded and server Dolt modes
**Fix**: Kill all dolt processes, use embedded mode only:
```bash
pkill -f "dolt sql-server"
lsof | grep "sfgastown.*dolt" | awk '{print $2}' | sort -u | xargs kill
```

### Issue: Plugin not copied, no errors shown

**Cause**: `rc.Hooks` is nil or `provider` is empty
**Fix**: Ensure settings have `runtime.provider: "opencode"`

### Issue: Binary reports "Built with go build directly"

**Cause**: Binary not built with `make build`
**Fix**: Rebuild properly:
```bash
cd ~/gt/sfgastown/crew/dunks
make install
```

## Files and References

**Source Code:**
- `internal/opencode/plugin.go` - Plugin copy logic with embed
- `internal/opencode/plugin/gastown.js` - Plugin source (embedded)
- `internal/runtime/runtime.go` - `EnsureSettingsForRole()` function
- `internal/config/types.go` - `RuntimeConfig`, `normalizeRuntimeConfig()`
- `internal/config/loader.go` - `ResolveRoleAgentConfig()`
- `internal/config/agents.go` - Agent presets including OpenCode
- `internal/polecat/session_manager.go` - Polecat creation, calls `EnsureSettingsForRole()`

**Settings:**
- `~/gt/settings/config.json` - Town-level settings
- `~/gt/<rig>/settings/config.json` - Rig-level settings

**Documentation:**
- `~/gt/docs/dolt.md` - Dolt backend documentation
- `~/gt/docs/migrating-to-dolt.md` - Dolt migration guide

## Summary

For OpenCode plugin to work correctly:

1. ✅ Settings must specify `runtime.provider: "opencode"`
2. ✅ Binary must be built with `make install` (includes embedded plugin)
3. ✅ `EnsureSettingsForRole()` must receive RuntimeConfig with Hooks configured
4. ✅ `opencode.EnsurePluginAt()` copies plugin to `.opencode/plugin/gastown.js`
5. ✅ Plugin runs `gt prime` on session start to inject context

**Current Status:** Settings are configured correctly, but binary may need rebuild.
