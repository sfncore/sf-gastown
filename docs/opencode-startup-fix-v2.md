# OpenCode Startup Fix v2: Root Cause Analysis & Implementation Plan

> **Issue**: OpenCode polecat agents intermittently fail to self-start (~40% failure rate).
> They sit at the "Ask anything" welcome screen forever, never receiving context or work.

## Root Cause

### The Two Config Normalization Paths

Gas Town has **two functions** that prepare `RuntimeConfig` for use:

| Function | Used by | Populates Tmux? | Populates Hooks? |
|----------|---------|-----------------|------------------|
| `normalizeRuntimeConfig()` (types.go:440) | Preset agents via `RuntimeConfigFromPreset` | **YES** (ReadyDelayMs, ProcessNames, etc.) | YES |
| `fillRuntimeDefaults()` (loader.go:1100) | Custom agents from town/rig settings | **NO** | YES (added in st-t1oi) |

Custom agents (like "Kimi K2.5" in our town settings) go through `fillRuntimeDefaults()`,
which was updated in st-t1oi to auto-fill `Hooks` defaults but **never received the
equivalent Tmux auto-fill**.

### The Config Chain

Town settings (`~/gt/settings/config.json`):
```json
{
  "Kimi K2.5": {
    "command": "opencode",
    "args": ["-m", "kimi-for-coding/kimi-k2.5"],
    "env": {"OPENCODE_PERMISSION": "{\"*\":\"allow\"}"},
    "prompt_mode": "none"
  }
}
```

After `fillRuntimeDefaults()`:
```
Provider:  "opencode"    ← auto-detected (st-t1oi fix) ✓
Hooks:     {Dir: ".opencode/plugin", SettingsFile: "gastown.js"}  ← auto-filled ✓
Tmux:      nil           ← NOT FILLED ✗  ← THE BUG
Session:   nil           ← NOT FILLED (minor, not critical)
```

After `normalizeRuntimeConfig()` (what SHOULD happen):
```
Tmux: {
  ReadyDelayMs:      8000,
  ProcessNames:      ["opencode", "node"],
  ReadyPromptPrefix: "",
}
```

### The Effect in Polecat Startup

In `session_manager.go`:
```go
Line 254: WaitForCommand(...)       // Waits for opencode process to appear
Line 260: SleepForReadyDelay(rc)    // rc.Tmux == nil → RETURNS IMMEDIATELY (should wait 8s)
Line 264: Combined nudge fires      // Fires 0ms after process detection
```

The nudge types beacon text into OpenCode's tmux pane **before the TUI is ready**.
Text is swallowed during TUI initialization. Agent never gets primed.

### Two Parallel Priming Mechanisms

**Mechanism A: gastown.js plugin**
- Deployed to `polecats/<name>/.opencode/plugin/gastown.js`
- OpenCode starts in `polecats/<name>/<rigname>/` (one dir deeper)
- IF OpenCode walks up dirs → plugin loads → `session.created` → `gt prime`
- Uncertain whether OpenCode walks up directories

**Mechanism B: tmux nudge**
- For `PromptMode: "none"` agents, beacon sent via `NudgeSession()`
- Designed to work WITH plugin (hook runs gt prime, nudge sends work instructions)
- `StartupNudgeDelayMs: 0` assumes hook already ran gt prime synchronously
- But without ReadyDelayMs, nudge fires before TUI is ready

### Why Intermittent (~40% failure)

- **Success**: OpenCode initializes fast → TUI ready before nudge → works.
  OR plugin loads and `session.created` fires independently of nudge timing.
- **Failure**: OpenCode slow to start → nudge text swallowed → plugin not found
  (CWD is one dir deeper) → agent stuck at welcome screen forever.

## Previous Fix Attempts

### st-t1oi: Provider Auto-Detection (MERGED)
- **What**: `fillRuntimeDefaults()` auto-detects `Provider` from command name
- **Fixed**: Provider defaulting to "claude" for custom opencode agents
- **Enabled**: Hooks auto-fill (Dir, SettingsFile) for custom agents
- **Missed**: Tmux auto-fill (ReadyDelayMs, ProcessNames)

### st-c2s.6: Town Settings Config (MERGED)
- **What**: Added `env`, `prompt_mode`, `process_names`, `non_interactive` to agents
- **Fixed**: Sessions dying on startup (no YOLO mode, wrong prompt mode)

### PR #978/#1029: Nudge Fallback System (MERGED)
- **What**: `StartupFallbackInfo` with beacon nudge + startup nudge
- **Fixed**: Non-hook/no-prompt agents getting no instructions
- **Depends on**: ReadyDelayMs being populated (which it isn't for custom agents)

## The Fix

### Primary: Tmux Auto-Fill in fillRuntimeDefaults

Add Tmux auto-fill block in `fillRuntimeDefaults()` (loader.go), after the existing
Hooks auto-fill at lines 1180-1195. Use the exact same pattern:

```go
// Auto-fill Tmux defaults based on provider.
// Custom agents need the same ReadyDelayMs and ProcessNames as preset agents.
if result.Tmux == nil {
    provider := result.Provider
    if provider == "" {
        provider = detectProviderFromCommand(result.Command)
    }
    tmux := &RuntimeTmuxConfig{
        ReadyDelayMs:      defaultReadyDelayMs(provider),
        ReadyPromptPrefix: defaultReadyPromptPrefix(provider),
        ProcessNames:      defaultProcessNames(provider, result.Command),
    }
    // Only set if we got meaningful defaults
    if tmux.ReadyDelayMs > 0 || len(tmux.ProcessNames) > 0 {
        result.Tmux = tmux
    }
}
```

### Secondary: Session Auto-Fill

Same pattern for Session (less critical but completes the fix):

```go
if result.Session == nil {
    provider := result.Provider
    sessionIDEnv := defaultSessionIDEnv(provider)
    configDirEnv := defaultConfigDirEnv(provider)
    if sessionIDEnv != "" || configDirEnv != "" {
        result.Session = &RuntimeSessionConfig{
            SessionIDEnv: sessionIDEnv,
            ConfigDirEnv: configDirEnv,
        }
    }
}
```

## Affected Startup Paths (ALL roles, not just polecat)

| Role | File | Has SleepForReadyDelay? | Has FallbackInfo nudge? | Has legacy RunStartupFallback? |
|------|------|:-:|:-:|:-:|
| **polecat** | `polecat/session_manager.go` | Yes (line 260) | Yes (line 192) | Yes (line 286) |
| **dog** | `dog/session_manager.go` | **NO** | **NO** | Yes (line 162) |
| **deacon** | `cmd/deacon.go` | **NO** | **NO** | Yes (line 461) |

Dog and deacon are missing BOTH the ready delay AND the newer fallback matrix.
They only have the old `RunStartupFallback` which returns nil when hooks are configured.

The fillRuntimeDefaults Tmux auto-fill (st-85iz) fixes ReadyDelayMs=0 for ALL paths.
The dog/deacon alignment task (st-z64r) adds SleepForReadyDelay + FallbackInfo nudges.

## Files to Modify

| File | Change |
|------|--------|
| `internal/config/loader.go` | Add Tmux + Session auto-fill in `fillRuntimeDefaults()` |
| `internal/config/loader_test.go` | Tests for Tmux auto-fill |
| `internal/doctor/` | New check: verify RuntimeConfig has Tmux for non-claude agents |
| `internal/dog/session_manager.go` | Add SleepForReadyDelay + FallbackInfo nudges |
| `internal/cmd/deacon.go` | Add SleepForReadyDelay + FallbackInfo nudges |
| `docs/opencode.md` | Update with config requirements |
| `settings/config.json` (town) | Example with explicit tmux config |

## Test Matrix

| Scenario | Expected ReadyDelayMs | Expected ProcessNames |
|----------|----------------------|----------------------|
| Custom opencode agent (no Tmux in config) | 8000 | ["opencode", "node"] |
| Custom claude agent (no Tmux in config) | 10000 | ["node", "claude"] |
| Custom agent with explicit Tmux | Preserved from config | Preserved from config |
| Preset opencode agent | 8000 | ["opencode", "node"] |
| Unknown command agent (no Tmux) | 0 | [command basename] |
| Nil RuntimeConfig | 10000 (claude default) | ["node", "claude"] |
