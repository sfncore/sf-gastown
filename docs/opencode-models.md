# OpenCode Model Configuration per Role

This document shows how to configure different AI models for different Gas Town roles when using OpenCode.

## Configuration Location

Settings are stored in:
- **Town-level**: `~/gt/settings/config.json`
- **Rig-level**: `~/gt/<rig>/settings/config.json`

## How It Works

OpenCode uses the `--model` flag with format `provider/model` (e.g., `anthropic/claude-3-5-sonnet-20241022`).

You define:
1. **Custom agents** in the `agents` field with specific models
2. **Role assignments** in the `role_agents` field mapping roles to agents

## Example Configuration

### Basic Setup (Town-Level)

```json
{
  "type": "town-settings",
  "version": 1,
  "default_agent": "opencode",
  "runtime": {
    "provider": "opencode"
  },
  "agents": {
    "opencode-claude-sonnet": {
      "command": "opencode",
      "args": ["--model", "anthropic/claude-3-5-sonnet-20241022"]
    },
    "opencode-claude-haiku": {
      "command": "opencode",
      "args": ["--model", "anthropic/claude-3-haiku-20240307"]
    },
    "opencode-gpt4o": {
      "command": "opencode",
      "args": ["--model", "openai/gpt-4o"]
    },
    "opencode-gpt4o-mini": {
      "command": "opencode",
      "args": ["--model", "openai/gpt-4o-mini"]
    }
  },
  "role_agents": {
    "mayor": "opencode-claude-sonnet",
    "deacon": "opencode-claude-haiku",
    "witness": "opencode-claude-haiku",
    "refinery": "opencode-claude-sonnet",
    "polecat": "opencode-gpt4o",
    "crew": "opencode-claude-sonnet"
  }
}
```

### Cost-Optimized Setup

Use cheaper models for background tasks:

```json
{
  "type": "town-settings",
  "version": 1,
  "default_agent": "opencode",
  "runtime": {
    "provider": "opencode"
  },
  "agents": {
    "opencode-smart": {
      "command": "opencode",
      "args": ["--model", "anthropic/claude-3-5-sonnet-20241022"]
    },
    "opencode-fast": {
      "command": "opencode",
      "args": ["--model", "anthropic/claude-3-haiku-20240307"]
    }
  },
  "role_agents": {
    "mayor": "opencode-smart",
    "deacon": "opencode-fast",
    "witness": "opencode-fast",
    "refinery": "opencode-smart",
    "polecat": "opencode-smart",
    "crew": "opencode-smart"
  }
}
```

### Rig-Level Override

Override town settings for a specific rig:

**File**: `~/gt/sfgastown/settings/config.json`

```json
{
  "type": "rig-settings",
  "version": 1,
  "runtime": {
    "provider": "opencode"
  },
  "agents": {
    "opencode-rig-specific": {
      "command": "opencode",
      "args": ["--model", "openai/gpt-4o"]
    }
  },
  "role_agents": {
    "polecat": "opencode-rig-specific"
  }
}
```

## Available Models

List available models with:
```bash
opencode models
```

Common providers:
- `anthropic/claude-3-5-sonnet-20241022`
- `anthropic/claude-3-opus-20240229`
- `anthropic/claude-3-haiku-20240307`
- `openai/gpt-4o`
- `openai/gpt-4o-mini`
- `openai/gpt-4-turbo`

## Role Descriptions

| Role | Typical Use | Recommended Model |
|------|-------------|-------------------|
| **mayor** | Global coordination, strategic decisions | Smart model (Sonnet/GPT-4) |
| **deacon** | Daemon, monitoring, heartbeats | Fast model (Haiku/GPT-4-mini) |
| **witness** | Polecat lifecycle management | Fast model (Haiku/GPT-4-mini) |
| **refinery** | Merge queue processing | Smart model (Sonnet/GPT-4) |
| **polecat** | Task execution, coding | Smart model (Sonnet/GPT-4) |
| **crew** | Human-managed workspace | User preference |

## Resolution Order

When a role spawns, the agent is selected in this order:

1. **Rig's `role_agents[role]`** - Rig-specific override
2. **Town's `role_agents[role]`** - Town-level assignment
3. **Rig's `Agent`** - Rig default
4. **Town's `default_agent`** - Town default
5. **"claude"** - Ultimate fallback

## Verification

Check which agent a role will use:
```bash
gt config agent list
```

Test polecat spawning:
```bash
gt sling <bead> <rig>
gt polecat status <rig>/<polecat>
```

## Current Settings

**Town-level** (`~/gt/settings/config.json`):
- Default agent: opencode
- Runtime provider: opencode
- No role-specific assignments (uses default for all)

**Rig-level** (`~/gt/sfgastown/settings/config.json`):
- Runtime provider: opencode
- No custom agents defined
- No role-specific assignments

To customize, add the `agents` and `role_agents` fields as shown in examples above.
