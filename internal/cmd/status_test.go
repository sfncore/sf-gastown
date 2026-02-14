package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/gastown/internal/beads"
	"github.com/steveyegge/gastown/internal/rig"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	_ = r.Close()

	return buf.String()
}

func TestDiscoverRigAgents_UsesRigPrefix(t *testing.T) {
	townRoot := t.TempDir()
	writeTestRoutes(t, townRoot, []beads.Route{
		{Prefix: "bd-", Path: "beads/mayor/rig"},
	})

	r := &rig.Rig{
		Name:       "beads",
		Path:       filepath.Join(townRoot, "beads"),
		HasWitness: true,
	}

	allAgentBeads := map[string]*beads.Issue{
		"bd-beads-witness": {
			ID:         "bd-beads-witness",
			AgentState: "running",
			HookBead:   "bd-hook",
		},
	}
	allHookBeads := map[string]*beads.Issue{
		"bd-hook": {ID: "bd-hook", Title: "Pinned"},
	}

	agents := discoverRigAgents(map[string]bool{}, r, nil, allAgentBeads, allHookBeads, nil, true)
	if len(agents) != 1 {
		t.Fatalf("discoverRigAgents() returned %d agents, want 1", len(agents))
	}

	if agents[0].State != "running" {
		t.Fatalf("agent state = %q, want %q", agents[0].State, "running")
	}
	if !agents[0].HasWork {
		t.Fatalf("agent HasWork = false, want true")
	}
	if agents[0].WorkTitle != "Pinned" {
		t.Fatalf("agent WorkTitle = %q, want %q", agents[0].WorkTitle, "Pinned")
	}
}

func TestRenderAgentDetails_UsesRigPrefix(t *testing.T) {
	townRoot := t.TempDir()
	writeTestRoutes(t, townRoot, []beads.Route{
		{Prefix: "bd-", Path: "beads/mayor/rig"},
	})

	agent := AgentRuntime{
		Name:    "witness",
		Address: "beads/witness",
		Role:    "witness",
		Running: true,
	}

	output := captureStdout(t, func() {
		renderAgentDetails(agent, "", nil, townRoot)
	})

	if !strings.Contains(output, "bd-beads-witness") {
		t.Fatalf("output %q does not contain rig-prefixed bead ID", output)
	}
}

func TestDiscoverRigAgents_ZombieSessionNotRunning(t *testing.T) {
	// Verify that a session in allSessions with value=false (zombie: tmux alive,
	// agent dead) results in agent.Running=false. This is the core fix for gt-bd6i3.
	townRoot := t.TempDir()
	writeTestRoutes(t, townRoot, []beads.Route{
		{Prefix: "gt-", Path: "gastown/mayor/rig"},
	})

	r := &rig.Rig{
		Name:       "gastown",
		Path:       filepath.Join(townRoot, "gastown"),
		HasWitness: true,
	}

	// allSessions has the witness session but marked as zombie (false).
	// This simulates a tmux session that exists but whose agent process has died.
	allSessions := map[string]bool{
		"gt-gastown-witness": false, // zombie: tmux exists, agent dead
	}

	agents := discoverRigAgents(allSessions, r, nil, nil, nil, nil, true)
	for _, a := range agents {
		if a.Role == "witness" {
			if a.Running {
				t.Fatal("zombie witness session (allSessions=false) should show as not running")
			}
			return
		}
	}
	t.Fatal("witness agent not found in results")
}

func TestDiscoverRigAgents_MissingSessionNotRunning(t *testing.T) {
	// Verify that a session not in allSessions at all results in agent.Running=false.
	townRoot := t.TempDir()
	writeTestRoutes(t, townRoot, []beads.Route{
		{Prefix: "gt-", Path: "gastown/mayor/rig"},
	})

	r := &rig.Rig{
		Name:       "gastown",
		Path:       filepath.Join(townRoot, "gastown"),
		HasWitness: true,
	}

	// Empty sessions map - no tmux sessions exist at all
	allSessions := map[string]bool{}

	agents := discoverRigAgents(allSessions, r, nil, nil, nil, nil, true)
	for _, a := range agents {
		if a.Role == "witness" {
			if a.Running {
				t.Fatal("witness with no tmux session should show as not running")
			}
			return
		}
	}
	t.Fatal("witness agent not found in results")
}

func TestBuildStatusIndicator_ZombieShowsStopped(t *testing.T) {
	// Verify that a zombie agent (Running=false) shows ○ (stopped), not ● (running)
	agent := AgentRuntime{Running: false}
	indicator := buildStatusIndicator(agent)
	if strings.Contains(indicator, "●") {
		t.Fatal("zombie agent (Running=false) should not show ● indicator")
	}
}

func TestBuildStatusIndicator_AliveShowsRunning(t *testing.T) {
	// Verify that an alive agent (Running=true) shows ● (running)
	agent := AgentRuntime{Running: true}
	indicator := buildStatusIndicator(agent)
	if strings.Contains(indicator, "○") {
		t.Fatal("alive agent (Running=true) should not show ○ indicator")
	}
}

func TestRunStatusWatch_RejectsZeroInterval(t *testing.T) {
	oldInterval := statusInterval
	oldWatch := statusWatch
	defer func() {
		statusInterval = oldInterval
		statusWatch = oldWatch
	}()

	statusInterval = 0
	statusWatch = true

	err := runStatusWatch(nil, nil)
	if err == nil {
		t.Fatal("expected error for zero interval, got nil")
	}
	if !strings.Contains(err.Error(), "positive") {
		t.Errorf("error %q should mention 'positive'", err.Error())
	}
}

func TestRunStatusWatch_RejectsNegativeInterval(t *testing.T) {
	oldInterval := statusInterval
	oldWatch := statusWatch
	defer func() {
		statusInterval = oldInterval
		statusWatch = oldWatch
	}()

	statusInterval = -5
	statusWatch = true

	err := runStatusWatch(nil, nil)
	if err == nil {
		t.Fatal("expected error for negative interval, got nil")
	}
	if !strings.Contains(err.Error(), "positive") {
		t.Errorf("error %q should mention 'positive'", err.Error())
	}
}

func TestRunStatusWatch_RejectsJSONCombo(t *testing.T) {
	oldJSON := statusJSON
	oldWatch := statusWatch
	oldInterval := statusInterval
	defer func() {
		statusJSON = oldJSON
		statusWatch = oldWatch
		statusInterval = oldInterval
	}()

	statusJSON = true
	statusWatch = true
	statusInterval = 2

	err := runStatusWatch(nil, nil)
	if err == nil {
		t.Fatal("expected error for --json + --watch, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be used together") {
		t.Errorf("error %q should mention 'cannot be used together'", err.Error())
	}
}

// TestParseRuntimeInfo tests the parseRuntimeInfo function with various cmdline patterns.
func TestParseRuntimeInfo(t *testing.T) {
	tests := []struct {
		name     string
		cmdline  string
		want     RuntimeInfo
	}{
		{
			name:    "empty cmdline",
			cmdline: "",
			want:    RuntimeInfo{},
		},
		{
			name:    "claude --model opus",
			cmdline: "claude --model opus",
			want:    RuntimeInfo{Provider: "claude", Model: "opus"},
		},
		{
			name:    "claude with short model flag",
			cmdline: "claude -m sonnet",
			want:    RuntimeInfo{Provider: "claude", Model: "sonnet"},
		},
		{
			name:    "pi direct command",
			cmdline: "pi --provider anthropic --model claude-sonnet",
			want:    RuntimeInfo{Provider: "pi", Model: "claude-sonnet"},
		},
		{
			name:    "pi with short flags",
			cmdline: "pi -p provider -m model-name",
			want:    RuntimeInfo{Provider: "pi", Model: "model-name"},
		},
		{
			name:    "opencode direct command",
			cmdline: "opencode --model gpt-4",
			want:    RuntimeInfo{Provider: "opencode", Model: "gpt-4"},
		},
		{
			name:    "node running pi wrapper",
			cmdline: "node /path/to/pi -e hooks.js",
			want:    RuntimeInfo{Provider: "pi", Model: ""},
		},
		{
			name:    "bun running opencode",
			cmdline: "bun run opencode",
			want:    RuntimeInfo{Provider: "opencode", Model: ""},
		},
		{
			name:    "node running pi with model",
			cmdline: "node /usr/local/bin/pi --model claude-opus",
			want:    RuntimeInfo{Provider: "pi", Model: "claude-opus"},
		},
		{
			name:    "gemini command",
			cmdline: "gemini --model flash",
			want:    RuntimeInfo{Provider: "gemini", Model: "flash"},
		},
		{
			name:    "codex command",
			cmdline: "codex --model o3-mini",
			want:    RuntimeInfo{Provider: "codex", Model: "o3-mini"},
		},
		{
			name:    "cursor-agent command",
			cmdline: "cursor-agent",
			want:    RuntimeInfo{Provider: "cursor", Model: ""},
		},
		{
			name:    "auggie command",
			cmdline: "auggie",
			want:    RuntimeInfo{Provider: "auggie", Model: ""},
		},
		{
			name:    "amp command",
			cmdline: "amp",
			want:    RuntimeInfo{Provider: "amp", Model: ""},
		},
		{
			name:    "claude with full path",
			cmdline: "/usr/local/bin/claude --model opus",
			want:    RuntimeInfo{Provider: "claude", Model: "opus"},
		},
		{
			name:    "unknown command",
			cmdline: "unknown-agent --model test",
			want:    RuntimeInfo{Provider: "", Model: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRuntimeInfo(tt.cmdline)
			if got.Provider != tt.want.Provider {
				t.Errorf("parseRuntimeInfo(%q).Provider = %q, want %q", tt.cmdline, got.Provider, tt.want.Provider)
			}
			if got.Model != tt.want.Model {
				t.Errorf("parseRuntimeInfo(%q).Model = %q, want %q", tt.cmdline, got.Model, tt.want.Model)
			}
		})
	}
}

// TestIsAgentCmdline tests the isAgentCmdline function for wrapper detection.
func TestIsAgentCmdline(t *testing.T) {
	tests := []struct {
		name    string
		cmdline string
		want    bool
	}{
		{
			name:    "empty cmdline",
			cmdline: "",
			want:    false,
		},
		{
			name:    "node running pi",
			cmdline: "node /path/to/pi -e hooks.js",
			want:    true,
		},
		{
			name:    "bun running pi",
			cmdline: "bun run pi",
			want:    true,
		},
		{
			name:    "node running opencode",
			cmdline: "node /path/to/opencode/cli.js",
			want:    true,
		},
		{
			name:    "bun running opencode",
			cmdline: "bun run opencode",
			want:    true,
		},
		{
			name:    "direct pi command",
			cmdline: "pi --model test",
			want:    false,
		},
		{
			name:    "direct opencode command",
			cmdline: "opencode",
			want:    false,
		},
		{
			name:    "node running other script",
			cmdline: "node /path/to/script.js",
			want:    false,
		},
		{
			name:    "bun running other script",
			cmdline: "bun run some-script",
			want:    false,
		},
		{
			name:    "claude command",
			cmdline: "claude --model opus",
			want:    false,
		},
		{
			name:    "node with pi in path",
			cmdline: "node /home/user/.local/bin/pi",
			want:    true,
		},
		{
			name:    "node with opencode in path",
			cmdline: "node /home/user/.local/bin/opencode",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAgentCmdline(tt.cmdline)
			if got != tt.want {
				t.Errorf("isAgentCmdline(%q) = %v, want %v", tt.cmdline, got, tt.want)
			}
		})
	}
}

// TestReadPiDefaults tests the readPiDefaults function.
func TestReadPiDefaults(t *testing.T) {
	t.Run("returns empty map when settings file does not exist", func(t *testing.T) {
		// Use a temp home directory to ensure no settings file exists
		oldHome := os.Getenv("HOME")
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", oldHome)

		settings, err := readPiDefaults()
		if err != nil {
			t.Fatalf("readPiDefaults() error = %v, want nil", err)
		}
		if settings == nil {
			t.Fatal("readPiDefaults() returned nil, want empty map")
		}
		if len(settings) != 0 {
			t.Errorf("readPiDefaults() returned %d entries, want 0", len(settings))
		}
	})

	t.Run("reads and parses settings file", func(t *testing.T) {
		oldHome := os.Getenv("HOME")
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", oldHome)

		// Create the settings directory and file
		settingsDir := filepath.Join(tmpHome, ".pi", "agent")
		if err := os.MkdirAll(settingsDir, 0755); err != nil {
			t.Fatalf("failed to create settings dir: %v", err)
		}

		settingsContent := `{
			"defaultProvider": "anthropic",
			"defaultModel": "claude-sonnet-4",
			"autoApprove": true
		}`
		settingsPath := filepath.Join(settingsDir, "settings.json")
		if err := os.WriteFile(settingsPath, []byte(settingsContent), 0644); err != nil {
			t.Fatalf("failed to write settings file: %v", err)
		}

		settings, err := readPiDefaults()
		if err != nil {
			t.Fatalf("readPiDefaults() error = %v, want nil", err)
		}

		if settings["defaultProvider"] != "anthropic" {
			t.Errorf("defaultProvider = %v, want anthropic", settings["defaultProvider"])
		}
		if settings["defaultModel"] != "claude-sonnet-4" {
			t.Errorf("defaultModel = %v, want claude-sonnet-4", settings["defaultModel"])
		}
		if autoApprove, ok := settings["autoApprove"].(bool); !ok || !autoApprove {
			t.Errorf("autoApprove = %v, want true", settings["autoApprove"])
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		oldHome := os.Getenv("HOME")
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", oldHome)

		// Create the settings directory and invalid file
		settingsDir := filepath.Join(tmpHome, ".pi", "agent")
		if err := os.MkdirAll(settingsDir, 0755); err != nil {
			t.Fatalf("failed to create settings dir: %v", err)
		}

		settingsPath := filepath.Join(settingsDir, "settings.json")
		if err := os.WriteFile(settingsPath, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("failed to write settings file: %v", err)
		}

		_, err := readPiDefaults()
		if err == nil {
			t.Error("readPiDefaults() error = nil, want error for invalid JSON")
		}
	})
}
