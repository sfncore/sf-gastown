# Pi Agent Polecat Behavior and GUPP Compliance

This document describes how pi agent polecats should behave in the Gas Town ecosystem, with specific focus on auto-execution (GUPP - "Go Unless Prompted to Pause"), hook detection, and session startup protocols.

## Overview

Pi agents in Gas Town operate as **polecats** - autonomous worker agents that execute work immediately upon startup without waiting for human confirmation. This behavior is fundamentally different from interactive pi usage and requires specific configuration and behavioral patterns.

## 1. Auto-Execution on Startup (GUPP)

### The Propulsion Principle

**"If you find work on your hook, YOU RUN IT."**

This is the core GUPP principle:
- No confirmation prompts
- No waiting for human input  
- No "announce and wait" behavior
- Work on the hook IS the assignment

### Why This Matters

Gas Town is designed as a steam engine where agents are pistons:
- There is no supervisor polling agents asking "did you start yet?"
- The hook IS the assignment - placed there deliberately
- Every moment of waiting stalls the entire system
- Other agents may be blocked waiting on YOUR output

### Failure Mode Prevention

The GUPP principle prevents this failure chain:
1. Agent starts with work on hook
2. Agent announces itself and waits for human "ok go"
3. Human is AFK / trusting the engine to run
4. Work sits idle → Gas Town stops

### Correct Startup Behavior

```typescript
// When pi agent starts as a polecat:
1. Check hook immediately (gt hook / bd show <issue>)
2. If work is hooked → EXECUTE immediately (no announcement, no waiting)
3. If hook empty → Check mail (gt mail inbox)
4. Still nothing? → Wait for user instructions
```

### Implementation Notes

Pi agents operating as polecats should:
- Use SDK `createAgentSession()` with appropriate configuration
- Set up session manager to check for hooked work before interactive mode
- Skip TUI initialization if work is found on hook
- Execute autonomously using the `runPrintMode` or custom execution loop

## 2. Hook Detection and Execution

### Hook Structure

In Gas Town, work is assigned via **hooks**:
- Each polecat has a hook where molecules (workflows) or mail can be attached
- Hooks are checked via `gt hook` or `bd show <bead-id>`
- Work MUST be hooked for polecats (they always have work)

### Hook Types

1. **Molecule (Workflow)**: A structured workflow with steps
   - Use `bd ready` to see workflow steps
   - Follow steps until completion
   - Mark steps complete with `bd close <step-id>`

2. **Mail**: Ad-hoc prose instructions
   - Read mail content
   - Interpret instructions
   - Execute directly

### Hook Detection Protocol

```bash
# Step 1: Check your hook
gt hook                          # Shows hooked work (if any)

# Step 2: Check your molecule steps  
bd ready                         # Shows your workflow steps - ALWAYS DO THIS

# Step 3: Work hooked? → Follow steps
# Hook empty? → Check mail for attached work
gt mail inbox

# Step 4: If mail contains attached work, hook it:
gt mol attach-from-mail <mail-id>

# Step 5: Execute from hook
gt prime                         # Load full context and begin
```

### SDK Implementation

For pi agent SDK integration:

```typescript
import { createAgentSession, runPrintMode } from "@mariozechner/pi-coding-agent";

// Check for hooked work before starting interactive mode
async function checkHookAndExecute() {
  // Pseudo-code for hook detection
  const hookedWork = await checkGasTownHook();
  
  if (hookedWork) {
    // Execute autonomously
    const { session } = await createAgentSession({
      sessionManager: SessionManager.inMemory(),
      // ... other config
    });
    
    await runPrintMode(session, {
      mode: "text",
      initialMessage: hookedWork.instructions,
    });
    
    // Signal completion
    await signalDone();
  } else {
    // Start interactive mode or wait
  }
}
```

## 3. Prompt Handling Differences from Claude

### Claude vs Pi Agent Differences

| Aspect | Claude Polecats | Pi Agent Polecats |
|--------|----------------|-------------------|
| **System Prompt** | Uses CLAUDE.md context files | Uses AGENTS.md or SYSTEM.md |
| **Tool Set** | Default: read, bash, edit, write | Same default set |
| **Session Management** | Persistent sessions | In-memory for polecats |
| **Startup** | Interactive by default | Must auto-detect hook |
| **Extensions** | TypeScript extensions | Same extension system |
| **Skills** | On-demand skill loading | Same skill system |

### Key Configuration Differences

**System Prompt Loading:**
- Pi loads `AGENTS.md` (or `CLAUDE.md` for compatibility) from:
  - `~/.pi/agent/AGENTS.md` (global)
  - Parent directories (walking up from cwd)
  - Current directory
- Can be overridden with `.pi/SYSTEM.md`

**Session Management:**
```typescript
// For polecats: use in-memory sessions
const { session } = await createAgentSession({
  sessionManager: SessionManager.inMemory(),
});

// Or ephemeral file-based
const { session } = await createAgentSession({
  sessionManager: SessionManager.create(process.cwd()),
});
```

**Tools:**
```typescript
// Default coding tools (same as Claude)
import { codingTools, readOnlyTools } from "@mariozechner/pi-coding-agent";

const { session } = await createAgentSession({
  tools: codingTools, // [read, bash, edit, write]
});
```

### Prompt Template Expansion

Pi supports prompt templates that expand via `/templatename`:

```markdown
<!-- ~/.pi/agent/prompts/gas-town-task.md -->
---
description: Execute a Gas Town task
---
You are a polecat in the Gas Town ecosystem. Execute the following task:

Task: $1
Priority: $2

Follow the Gas Town workflow:
1. Check your hook
2. Execute immediately without confirmation
3. Signal completion with gt done
```

## 4. Session Startup Protocol

### Complete Startup Sequence

```typescript
import { 
  createAgentSession, 
  SessionManager,
  SettingsManager,
  AuthStorage,
  ModelRegistry,
  runPrintMode 
} from "@mariozechner/pi-coding-agent";

async function polecatStartup() {
  // 1. Check for hooked work (GUPP)
  const hookedWork = await checkGasTownHook();
  
  if (!hookedWork) {
    // No work - exit or wait
    console.log("No work on hook. Exiting.");
    return;
  }
  
  // 2. Set up auth and models
  const authStorage = new AuthStorage();
  const modelRegistry = new ModelRegistry(authStorage);
  
  // 3. Create session (in-memory for polecats)
  const { session } = await createAgentSession({
    sessionManager: SessionManager.inMemory(),
    authStorage,
    modelRegistry,
    settingsManager: SettingsManager.inMemory({
      compaction: { enabled: false }, // Disable for short-lived sessions
    }),
  });
  
  // 4. Subscribe to events for logging
  session.subscribe((event) => {
    if (event.type === "message_update") {
      // Log or stream output
    }
  });
  
  // 5. Execute the work
  await runPrintMode(session, {
    mode: "text",
    initialMessage: hookedWork.instructions,
  });
  
  // 6. Signal completion to Gas Town
  await signalDone();
}
```

### Environment Setup

**Required Environment Variables:**
```bash
# API keys (provider-specific)
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-...

# Pi configuration
export PI_CODING_AGENT_DIR="~/.pi/agent"
```

**Context Files:**
Place `AGENTS.md` in the polecat's working directory:
```markdown
# Gas Town Polecat Instructions

You are a polecat in the Gas Town ecosystem.

## Behavior
- Execute work immediately upon finding it on your hook
- Do not wait for confirmation
- Do not announce your intentions beyond minimal logging
- Signal completion with `gt done`

## Gas Town Commands
- `gt hook` - Check for work
- `bd ready` - See workflow steps
- `bd close <id>` - Mark step complete
- `gt done` - Signal work complete
```

## 5. Completion Protocol

### Pre-Submission Checklist

Before signaling completion:

```bash
# 1. Check git status - must be clean
git status

# 2. Verify commits are present
git log --oneline -3

# 3. Push to remote
git push
```

### Signaling Done

```bash
# Submit work to merge queue
gt done
```

This single command:
- Verifies git is clean
- Submits branch to merge queue
- Handles beads sync internally

**Critical:** Do NOT manually close the root issue with `bd close`. The Refinery closes it after successful merge.

## 6. Error Handling and Escalation

### When to Escalate

Escalate when:
- Requirements are unclear after checking docs
- Stuck for >15 minutes on the same problem
- Found something blocking but outside your scope
- Tests fail and can't determine why after 2-3 attempts
- Need credentials, secrets, or external access

### Escalation Methods

```bash
# Option 1: gt escalate (preferred)
gt escalate "Brief description" -s HIGH -m "Details..."

# Option 2: Mail the Witness
gt mail send sfgastown/witness -s "HELP: problem" -m "Details..."

# Option 3: Mail the Mayor
gt mail send mayor/ -s "BLOCKED: topic" -m "Details..."
```

### After Escalating

1. If you can continue with other work → continue
2. If completely blocked → `gt done --status=ESCALATED`
3. Do NOT sit idle waiting for response

## 7. Key Differences Summary

| Feature | Interactive Pi | Pi Polecat |
|---------|---------------|------------|
| **Startup** | TUI interactive | Auto-execute from hook |
| **Session** | Persistent file | In-memory or ephemeral |
| **Input** | User prompts | Hooked work instructions |
| **Output** | TUI display | Log + git commits |
| **Completion** | User continues | `gt done` signal |
| **Confirmation** | Always | Never (GUPP) |
| **Error handling** | User resolves | Escalate and exit |

## References

- [Pi SDK Documentation](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent/docs/sdk.md)
- [Pi Extensions](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent/docs/extensions.md)
- [Pi Session Format](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent/docs/session.md)
- [Gas Town AGENTS.md](../AGENTS.md)
