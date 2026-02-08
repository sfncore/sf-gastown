package doctor

import (
	"testing"

	"github.com/sfncore/sf-gastown/internal/config"
)

func TestAgentTmuxConfigCheck_Name(t *testing.T) {
	c := NewAgentTmuxConfigCheck()
	if got := c.Name(); got != "agent-tmux-config" {
		t.Errorf("Name() = %v, want %v", got, "agent-tmux-config")
	}
}

func TestAgentTmuxConfigCheck_Description(t *testing.T) {
	c := NewAgentTmuxConfigCheck()
	if got := c.Description(); got != "Verify all role agents have proper Tmux configuration" {
		t.Errorf("Description() = %v, want %v", got, "Verify all role agents have proper Tmux configuration")
	}
}

func TestAgentTmuxConfigCheck_Category(t *testing.T) {
	c := NewAgentTmuxConfigCheck()
	if got := c.Category(); got != CategoryConfig {
		t.Errorf("Category() = %v, want %v", got, CategoryConfig)
	}
}

func TestAgentTmuxConfigCheck_CanFix(t *testing.T) {
	c := NewAgentTmuxConfigCheck()
	if c.CanFix() {
		t.Error("CanFix() = true, want false")
	}
}

func TestAgentTmuxConfigCheck_checkTmuxConfig(t *testing.T) {
	c := NewAgentTmuxConfigCheck()

	tests := []struct {
		name      string
		role      string
		agentName string
		rc        *config.RuntimeConfig
		wantIssue bool
	}{
		{
			name:      "valid claude config",
			role:      "mayor",
			agentName: "claude",
			rc: &config.RuntimeConfig{
				Tmux: &config.RuntimeTmuxConfig{
					ReadyDelayMs: 8000,
					ProcessNames: []string{"claude"},
				},
			},
			wantIssue: false,
		},
		{
			name:      "valid opencode config",
			role:      "polecat",
			agentName: "opencode",
			rc: &config.RuntimeConfig{
				Tmux: &config.RuntimeTmuxConfig{
					ReadyDelayMs: 8000,
					ProcessNames: []string{"opencode"},
				},
			},
			wantIssue: false,
		},
		{
			name:      "nil tmux config",
			role:      "witness",
			agentName: "my-agent",
			rc: &config.RuntimeConfig{
				Tmux: nil,
			},
			wantIssue: true,
		},
		{
			name:      "opencode with zero ReadyDelayMs",
			role:      "polecat",
			agentName: "opencode",
			rc: &config.RuntimeConfig{
				Tmux: &config.RuntimeTmuxConfig{
					ReadyDelayMs: 0,
					ProcessNames: []string{"opencode"},
				},
			},
			wantIssue: true,
		},
		{
			name:      "empty ProcessNames",
			role:      "crew",
			agentName: "claude",
			rc: &config.RuntimeConfig{
				Tmux: &config.RuntimeTmuxConfig{
					ReadyDelayMs: 8000,
					ProcessNames: []string{},
				},
			},
			wantIssue: true,
		},
		{
			name:      "gemini config without delay (OK)",
			role:      "refinery",
			agentName: "gemini",
			rc: &config.RuntimeConfig{
				Tmux: &config.RuntimeTmuxConfig{
					ReadyDelayMs: 0,
					ProcessNames: []string{"gemini"},
				},
			},
			wantIssue: false,
		},
		{
			name:      "codex with command containing claude (needs delay)",
			role:      "witness",
			agentName: "custom-codex",
			rc: &config.RuntimeConfig{
				Command: "claude-code",
				Tmux: &config.RuntimeTmuxConfig{
					ReadyDelayMs: 0,
					ProcessNames: []string{"claude-code"},
				},
			},
			wantIssue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.checkTmuxConfig(tt.role, tt.agentName, tt.rc)
			if tt.wantIssue && got == nil {
				t.Errorf("checkTmuxConfig() = nil, want issue")
			}
			if !tt.wantIssue && got != nil {
				t.Errorf("checkTmuxConfig() = %v, want nil", got)
			}
		})
	}
}
