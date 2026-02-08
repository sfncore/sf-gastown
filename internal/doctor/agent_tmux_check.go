package doctor

import (
	"fmt"
	"os"
	"strings"

	"github.com/sfncore/sf-gastown/internal/config"
)

// AgentTmuxConfigCheck verifies that all configured agents have proper Tmux defaults.
// This catches misconfigurations that would cause startup failures.
type AgentTmuxConfigCheck struct {
	BaseCheck
}

// tmuxIssue represents a detected Tmux configuration issue.
type tmuxIssue struct {
	role       string
	agentName  string
	problem    string
	suggestion string
}

// NewAgentTmuxConfigCheck creates a new agent Tmux config validation check.
func NewAgentTmuxConfigCheck() *AgentTmuxConfigCheck {
	return &AgentTmuxConfigCheck{
		BaseCheck: BaseCheck{
			CheckName:        "agent-tmux-config",
			CheckDescription: "Verify all role agents have proper Tmux configuration",
			CheckCategory:    CategoryConfig,
		},
	}
}

// Run checks all role_agents configurations for proper Tmux settings.
func (c *AgentTmuxConfigCheck) Run(ctx *CheckContext) *CheckResult {
	var issues []tmuxIssue
	var details []string

	// Load town settings
	townSettings, err := config.LoadOrCreateTownSettings(config.TownSettingsPath(ctx.TownRoot))
	if err != nil {
		townSettings = config.NewTownSettings()
	}

	// Collect all unique roles from town and rig settings
	rolesToCheck := make(map[string]bool)

	// Add town-level role_agents
	for role := range townSettings.RoleAgents {
		rolesToCheck[role] = true
	}

	// Add rig-level role_agents if checking a specific rig
	var rigSettings *config.RigSettings
	if ctx.RigName != "" {
		rigPath := ctx.RigPath()
		rigSettings, _ = config.LoadRigSettings(config.RigSettingsPath(rigPath))
		if rigSettings != nil {
			for role := range rigSettings.RoleAgents {
				rolesToCheck[role] = true
			}
		}
	}

	// Also check all known roles even if not in role_agents (uses defaults)
	knownRoles := []string{"mayor", "deacon", "witness", "refinery", "polecat", "crew", "dog"}
	for _, role := range knownRoles {
		rolesToCheck[role] = true
	}

	// Check each role
	for role := range rolesToCheck {
		var rigPath string
		if ctx.RigName != "" {
			rigPath = ctx.RigPath()
		}

		// Get the runtime config for this role
		rc := config.ResolveRoleAgentConfig(role, ctx.TownRoot, rigPath)
		if rc == nil {
			continue
		}

		// Get agent name for display
		agentName, _ := config.ResolveRoleAgentName(role, ctx.TownRoot, rigPath)

		// Check Tmux configuration
		if issue := c.checkTmuxConfig(role, agentName, rc); issue != nil {
			issues = append(issues, *issue)
			details = append(details, fmt.Sprintf("role_agents[%s] (%s): %s", role, agentName, issue.problem))
		}
	}

	if len(issues) == 0 {
		return &CheckResult{
			Name:    c.Name(),
			Status:  StatusOK,
			Message: "All role agents have proper Tmux configuration",
		}
	}

	return &CheckResult{
		Name:    c.Name(),
		Status:  StatusWarning,
		Message: fmt.Sprintf("Found %d role agent(s) with missing Tmux configuration", len(issues)),
		Details: details,
		FixHint: "Rebuild binary with fillRuntimeDefaults to auto-populate Tmux defaults",
	}
}

// checkTmuxConfig validates the Tmux configuration for a specific role.
// Returns nil if OK, otherwise returns an issue description.
func (c *AgentTmuxConfigCheck) checkTmuxConfig(role, agentName string, rc *config.RuntimeConfig) *tmuxIssue {
	// Check if Tmux config is nil
	if rc.Tmux == nil {
		return &tmuxIssue{
			role:       role,
			agentName:  agentName,
			problem:    "Tmux config is nil",
			suggestion: "RuntimeConfig.Tmux should be auto-populated by fillRuntimeDefaults",
		}
	}

	// Agents that need ReadyDelayMs for proper startup detection
	// These agents use prompt-based or delay-based readiness detection
	agentsNeedingDelay := []string{"opencode", "claude", "codex"}
	needsDelay := false
	for _, agent := range agentsNeedingDelay {
		if strings.Contains(strings.ToLower(agentName), agent) {
			needsDelay = true
			break
		}
	}

	// Also check by command name if agent name doesn't match
	if !needsDelay && rc.Command != "" {
		cmd := strings.ToLower(rc.Command)
		for _, agent := range agentsNeedingDelay {
			if strings.Contains(cmd, agent) {
				needsDelay = true
				break
			}
		}
	}

	if needsDelay && rc.Tmux.ReadyDelayMs <= 0 {
		return &tmuxIssue{
			role:       role,
			agentName:  agentName,
			problem:    fmt.Sprintf("Tmux.ReadyDelayMs is %d (nudge will fire too early)", rc.Tmux.ReadyDelayMs),
			suggestion: "Add tmux config to agent definition or use a known command name",
		}
	}

	// Check ProcessNames is not empty for agents that need it
	if len(rc.Tmux.ProcessNames) == 0 {
		return &tmuxIssue{
			role:       role,
			agentName:  agentName,
			problem:    "Tmux.ProcessNames is empty",
			suggestion: "RuntimeConfig.Tmux.ProcessNames should be auto-populated by fillRuntimeDefaults",
		}
	}

	return nil
}

// Fix returns an error since this check cannot be auto-fixed.
// The fix requires rebuilding the binary with updated fillRuntimeDefaults.
func (c *AgentTmuxConfigCheck) Fix(ctx *CheckContext) error {
	fmt.Fprintf(os.Stderr, "\n  Note: Tmux configuration issues are resolved by rebuilding the binary\n")
	fmt.Fprintf(os.Stderr, "  with the fillRuntimeDefaults fix. This cannot be auto-fixed at runtime.\n\n")
	return ErrCannotFix
}
