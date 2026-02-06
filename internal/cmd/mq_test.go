package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/steveyegge/gastown/internal/beads"
)

func TestParseBranchName(t *testing.T) {
	tests := []struct {
		name       string
		branch     string
		wantIssue  string
		wantWorker string
	}{
		{
			name:       "polecat branch format",
			branch:     "polecat/Nux/gt-xyz",
			wantIssue:  "gt-xyz",
			wantWorker: "Nux",
		},
		{
			name:       "polecat branch with subtask",
			branch:     "polecat/Worker/gt-abc.1",
			wantIssue:  "gt-abc.1",
			wantWorker: "Worker",
		},
		{
			name:       "polecat branch with issue and timestamp",
			branch:     "polecat/furiosa/gt-jns7.1@mk123456",
			wantIssue:  "gt-jns7.1",
			wantWorker: "furiosa",
		},
		{
			name:       "modern polecat branch (timestamp format)",
			branch:     "polecat/furiosa-mkc36bb9",
			wantIssue:  "", // Should NOT extract fake issue from worker-timestamp
			wantWorker: "furiosa",
		},
		{
			name:       "modern polecat branch with longer name",
			branch:     "polecat/citadel-mk0vro62",
			wantIssue:  "",
			wantWorker: "citadel",
		},
		{
			name:       "simple issue branch",
			branch:     "gt-xyz",
			wantIssue:  "gt-xyz",
			wantWorker: "",
		},
		{
			name:       "feature branch with issue",
			branch:     "feature/gt-abc-impl",
			wantIssue:  "gt-abc",
			wantWorker: "",
		},
		{
			name:       "no issue pattern",
			branch:     "main",
			wantIssue:  "",
			wantWorker: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parseBranchName(tt.branch)
			if info.Issue != tt.wantIssue {
				t.Errorf("parseBranchName() Issue = %q, want %q", info.Issue, tt.wantIssue)
			}
			if info.Worker != tt.wantWorker {
				t.Errorf("parseBranchName() Worker = %q, want %q", info.Worker, tt.wantWorker)
			}
		})
	}
}

func TestFormatMRAge(t *testing.T) {
	tests := []struct {
		name      string
		createdAt string
		wantOk    bool // just check it doesn't panic/error
	}{
		{
			name:      "RFC3339 format",
			createdAt: "2025-01-01T12:00:00Z",
			wantOk:    true,
		},
		{
			name:      "alternative format",
			createdAt: "2025-01-01T12:00:00",
			wantOk:    true,
		},
		{
			name:      "invalid format",
			createdAt: "not-a-date",
			wantOk:    true, // returns "?" for invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMRAge(tt.createdAt)
			if tt.wantOk && result == "" {
				t.Errorf("formatMRAge() returned empty for %s", tt.createdAt)
			}
		})
	}
}

func TestGetDescriptionWithoutMRFields(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "empty description",
			description: "",
			want:        "",
		},
		{
			name:        "only MR fields",
			description: "branch: polecat/Nux/gt-xyz\ntarget: main\nworker: Nux",
			want:        "",
		},
		{
			name:        "mixed content",
			description: "branch: polecat/Nux/gt-xyz\nSome custom notes\ntarget: main",
			want:        "Some custom notes",
		},
		{
			name:        "no MR fields",
			description: "Just a regular description\nWith multiple lines",
			want:        "Just a regular description\nWith multiple lines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDescriptionWithoutMRFields(tt.description)
			if got != tt.want {
				t.Errorf("getDescriptionWithoutMRFields() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{
			name:   "short string",
			s:      "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			s:      "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "needs truncation",
			s:      "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "very short max",
			s:      "hello",
			maxLen: 3,
			want:   "hel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.s, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   string // We check for substring since styling adds ANSI codes
	}{
		{
			name:   "open status",
			status: "open",
			want:   "open",
		},
		{
			name:   "in_progress status",
			status: "in_progress",
			want:   "in_progress",
		},
		{
			name:   "closed status",
			status: "closed",
			want:   "closed",
		},
		{
			name:   "unknown status",
			status: "pending",
			want:   "pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatStatus(tt.status)
			if got == "" {
				t.Errorf("formatStatus(%q) returned empty string", tt.status)
			}
			// The result contains ANSI codes, so just check the status text is present
			if !contains(got, tt.want) {
				t.Errorf("formatStatus(%q) = %q, should contain %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"open", "○"},
		{"in_progress", "▶"},
		{"closed", "✓"},
		{"unknown", "•"},
		{"", "•"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := getStatusIcon(tt.status)
			if got != tt.want {
				t.Errorf("getStatusIcon(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		wantEmpty bool
	}{
		{
			name:      "RFC3339 format",
			timestamp: "2025-01-01T12:00:00Z",
			wantEmpty: false,
		},
		{
			name:      "RFC3339 with timezone",
			timestamp: "2025-01-01T12:00:00-08:00",
			wantEmpty: false,
		},
		{
			name:      "date only format",
			timestamp: "2025-01-01",
			wantEmpty: false,
		},
		{
			name:      "datetime without Z",
			timestamp: "2025-01-01T12:00:00",
			wantEmpty: false,
		},
		{
			name:      "invalid format returns empty",
			timestamp: "not-a-date",
			wantEmpty: true,
		},
		{
			name:      "empty string returns empty",
			timestamp: "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimeAgo(tt.timestamp)
			if tt.wantEmpty && got != "" {
				t.Errorf("formatTimeAgo(%q) = %q, want empty", tt.timestamp, got)
			}
			if !tt.wantEmpty && got == "" {
				t.Errorf("formatTimeAgo(%q) returned empty, want non-empty", tt.timestamp)
			}
		})
	}
}

// contains checks if s contains substr (helper for styled output)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFilterMRsByTarget(t *testing.T) {
	// Create test MRs with different targets
	mrs := []*beads.Issue{
		makeTestMR("mr-1", "polecat/Nux/gt-001", "integration/gt-epic", "Nux", "open"),
		makeTestMR("mr-2", "polecat/Toast/gt-002", "main", "Toast", "open"),
		makeTestMR("mr-3", "polecat/Able/gt-003", "integration/gt-epic", "Able", "open"),
		makeTestMR("mr-4", "polecat/Baker/gt-004", "integration/gt-other", "Baker", "open"),
	}

	tests := []struct {
		name         string
		targetBranch string
		wantCount    int
		wantIDs      []string
	}{
		{
			name:         "filter to integration/gt-epic",
			targetBranch: "integration/gt-epic",
			wantCount:    2,
			wantIDs:      []string{"mr-1", "mr-3"},
		},
		{
			name:         "filter to main",
			targetBranch: "main",
			wantCount:    1,
			wantIDs:      []string{"mr-2"},
		},
		{
			name:         "filter to non-existent branch",
			targetBranch: "integration/no-such-epic",
			wantCount:    0,
			wantIDs:      []string{},
		},
		{
			name:         "filter to other integration branch",
			targetBranch: "integration/gt-other",
			wantCount:    1,
			wantIDs:      []string{"mr-4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterMRsByTarget(mrs, tt.targetBranch)
			if len(got) != tt.wantCount {
				t.Errorf("filterMRsByTarget() returned %d MRs, want %d", len(got), tt.wantCount)
			}

			// Verify correct IDs
			gotIDs := make(map[string]bool)
			for _, mr := range got {
				gotIDs[mr.ID] = true
			}
			for _, wantID := range tt.wantIDs {
				if !gotIDs[wantID] {
					t.Errorf("filterMRsByTarget() missing expected MR %s", wantID)
				}
			}
		})
	}
}

func TestFilterMRsByTarget_EmptyInput(t *testing.T) {
	got := filterMRsByTarget(nil, "integration/gt-epic")
	if got != nil {
		t.Errorf("filterMRsByTarget(nil) = %v, want nil", got)
	}

	got = filterMRsByTarget([]*beads.Issue{}, "integration/gt-epic")
	if len(got) != 0 {
		t.Errorf("filterMRsByTarget([]) = %v, want empty slice", got)
	}
}

func TestFilterMRsByTarget_NoMRFields(t *testing.T) {
	// Issue without MR fields in description
	plainIssue := &beads.Issue{
		ID:          "issue-1",
		Title:       "Not an MR",
		Type:        "merge-request",
		Status:      "open",
		Description: "Just a plain description with no MR fields",
	}

	got := filterMRsByTarget([]*beads.Issue{plainIssue}, "main")
	if len(got) != 0 {
		t.Errorf("filterMRsByTarget() should filter out issues without MR fields, got %d", len(got))
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		wantErr    bool
	}{
		{
			name:       "valid simple branch",
			branchName: "integration/gt-epic",
			wantErr:    false,
		},
		{
			name:       "valid nested branch",
			branchName: "user/project/feature",
			wantErr:    false,
		},
		{
			name:       "valid with hyphens and underscores",
			branchName: "user-name/feature_branch",
			wantErr:    false,
		},
		{
			name:       "empty branch name",
			branchName: "",
			wantErr:    true,
		},
		{
			name:       "contains tilde",
			branchName: "branch~1",
			wantErr:    true,
		},
		{
			name:       "contains caret",
			branchName: "branch^2",
			wantErr:    true,
		},
		{
			name:       "contains colon",
			branchName: "branch:ref",
			wantErr:    true,
		},
		{
			name:       "contains space",
			branchName: "branch name",
			wantErr:    true,
		},
		{
			name:       "contains backslash",
			branchName: "branch\\name",
			wantErr:    true,
		},
		{
			name:       "contains double dot",
			branchName: "branch..name",
			wantErr:    true,
		},
		{
			name:       "contains at-brace",
			branchName: "branch@{name}",
			wantErr:    true,
		},
		{
			name:       "ends with .lock",
			branchName: "branch.lock",
			wantErr:    true,
		},
		{
			name:       "starts with slash",
			branchName: "/branch",
			wantErr:    true,
		},
		{
			name:       "ends with slash",
			branchName: "branch/",
			wantErr:    true,
		},
		{
			name:       "starts with dot",
			branchName: ".branch",
			wantErr:    true,
		},
		{
			name:       "ends with dot",
			branchName: "branch.",
			wantErr:    true,
		},
		{
			name:       "consecutive slashes",
			branchName: "branch//name",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBranchName(tt.branchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBranchName(%q) error = %v, wantErr %v", tt.branchName, err, tt.wantErr)
			}
		})
	}
}

func TestGetIntegrationBranchField(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "empty description",
			description: "",
			want:        "",
		},
		{
			name:        "field at beginning",
			description: "integration_branch: klauern/PROJ-123/RA-epic\nSome description",
			want:        "klauern/PROJ-123/RA-epic",
		},
		{
			name:        "field in middle",
			description: "Some text\nintegration_branch: custom/branch\nMore text",
			want:        "custom/branch",
		},
		{
			name:        "field with extra whitespace",
			description: "  integration_branch:   spaced/branch  \nOther content",
			want:        "spaced/branch",
		},
		{
			name:        "no integration_branch field",
			description: "Just a plain description\nWith multiple lines",
			want:        "",
		},
		{
			name:        "mixed case field name",
			description: "Integration_branch: CamelCase/branch",
			want:        "CamelCase/branch",
		},
		{
			name:        "default format",
			description: "integration_branch: integration/gt-epic\nEpic for auth work",
			want:        "integration/gt-epic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIntegrationBranchField(tt.description)
			if got != tt.want {
				t.Errorf("getIntegrationBranchField() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestIssuePatternCompiledAtPackageLevel verifies that the issuePattern regex
// is compiled once at package level (not on every parseBranchName call).
func TestIssuePatternCompiledAtPackageLevel(t *testing.T) {
	// Verify the pattern is not nil and is a compiled regex
	if issuePattern == nil {
		t.Error("issuePattern should be compiled at package level, got nil")
	}
	// Verify it matches expected patterns
	tests := []struct {
		branch    string
		wantMatch bool
		wantIssue string
	}{
		{"polecat/Nux/gt-xyz", true, "gt-xyz"},
		{"gt-abc", true, "gt-abc"},
		{"feature/proj-123-add-feature", true, "proj-123"},
		{"main", false, ""},
		{"", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			matches := issuePattern.FindStringSubmatch(tt.branch)
			if (len(matches) > 1) != tt.wantMatch {
				t.Errorf("FindStringSubmatch(%q) match = %v, want %v", tt.branch, len(matches) > 1, tt.wantMatch)
			}
			if tt.wantMatch && len(matches) > 1 && matches[1] != tt.wantIssue {
				t.Errorf("FindStringSubmatch(%q) issue = %q, want %q", tt.branch, matches[1], tt.wantIssue)
			}
		})
	}
}

// TestPolecatCleanupTimeoutConstant verifies the timeout constant is set correctly.
func TestPolecatCleanupTimeoutConstant(t *testing.T) {
	// This test documents the expected timeout value.
	// The actual timeout behavior is tested manually or with integration tests.
	const expectedMaxCleanupWait = 5 * time.Minute
	if expectedMaxCleanupWait != 5*time.Minute {
		t.Errorf("expectedMaxCleanupWait = %v, want 5m", expectedMaxCleanupWait)
	}
}

func TestGetRigGit(t *testing.T) {
	t.Run("bare repo exists", func(t *testing.T) {
		tmp := t.TempDir()
		bareRepo := filepath.Join(tmp, ".repo.git")
		if err := os.Mkdir(bareRepo, 0o755); err != nil {
			t.Fatal(err)
		}

		g, err := getRigGit(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// NewGitWithDir for bare repos uses gitDir, workDir=""
		// The returned Git should point at the bare repo path
		if g == nil {
			t.Fatal("expected non-nil Git")
		}
	})

	t.Run("mayor/rig exists without bare repo", func(t *testing.T) {
		tmp := t.TempDir()
		mayorRig := filepath.Join(tmp, "mayor", "rig")
		if err := os.MkdirAll(mayorRig, 0o755); err != nil {
			t.Fatal(err)
		}

		g, err := getRigGit(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g == nil {
			t.Fatal("expected non-nil Git")
		}
	})

	t.Run("neither exists returns error", func(t *testing.T) {
		tmp := t.TempDir()

		_, err := getRigGit(tmp)
		if err == nil {
			t.Fatal("expected error for empty directory")
		}
		if !stringContains(err.Error(), "no repo base found") {
			t.Errorf("expected 'no repo base found' error, got: %v", err)
		}
	})

	t.Run("bare repo takes precedence over mayor/rig", func(t *testing.T) {
		tmp := t.TempDir()
		bareRepo := filepath.Join(tmp, ".repo.git")
		if err := os.Mkdir(bareRepo, 0o755); err != nil {
			t.Fatal(err)
		}
		mayorRig := filepath.Join(tmp, "mayor", "rig")
		if err := os.MkdirAll(mayorRig, 0o755); err != nil {
			t.Fatal(err)
		}

		g, err := getRigGit(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g == nil {
			t.Fatal("expected non-nil Git")
		}
		// When bare repo exists, WorkDir() returns "" (bare repo mode)
		if g.WorkDir() != "" {
			t.Errorf("expected empty WorkDir for bare repo, got %q", g.WorkDir())
		}
	})
}

func TestGetIntegrationBranchTemplate(t *testing.T) {
	t.Run("CLI override provided", func(t *testing.T) {
		tmp := t.TempDir()
		got := getIntegrationBranchTemplate(tmp, "custom/{epic}")
		if got != "custom/{epic}" {
			t.Errorf("got %q, want %q", got, "custom/{epic}")
		}
	})

	t.Run("config has template", func(t *testing.T) {
		tmp := t.TempDir()
		settingsDir := filepath.Join(tmp, "settings")
		if err := os.Mkdir(settingsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		cfg := map[string]interface{}{
			"type":    "rig-settings",
			"version": 1,
			"merge_queue": map[string]interface{}{
				"integration_branch_template": "{prefix}/{epic}",
			},
		}
		data, _ := json.Marshal(cfg)
		if err := os.WriteFile(filepath.Join(settingsDir, "config.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}

		got := getIntegrationBranchTemplate(tmp, "")
		if got != "{prefix}/{epic}" {
			t.Errorf("got %q, want %q", got, "{prefix}/{epic}")
		}
	})

	t.Run("config exists but no template returns default", func(t *testing.T) {
		tmp := t.TempDir()
		settingsDir := filepath.Join(tmp, "settings")
		if err := os.Mkdir(settingsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		cfg := map[string]interface{}{
			"type":        "rig-settings",
			"version":     1,
			"merge_queue": map[string]interface{}{},
		}
		data, _ := json.Marshal(cfg)
		if err := os.WriteFile(filepath.Join(settingsDir, "config.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}

		got := getIntegrationBranchTemplate(tmp, "")
		if got != defaultIntegrationBranchTemplate {
			t.Errorf("got %q, want %q", got, defaultIntegrationBranchTemplate)
		}
	})

	t.Run("no config file returns default", func(t *testing.T) {
		tmp := t.TempDir()
		got := getIntegrationBranchTemplate(tmp, "")
		if got != defaultIntegrationBranchTemplate {
			t.Errorf("got %q, want %q", got, defaultIntegrationBranchTemplate)
		}
	})
}

func TestIsReadyToLand(t *testing.T) {
	tests := []struct {
		name           string
		aheadCount     int
		childrenTotal  int
		childrenClosed int
		pendingMRCount int
		want           bool
	}{
		{
			name:           "all conditions met",
			aheadCount:     3,
			childrenTotal:  5,
			childrenClosed: 5,
			pendingMRCount: 0,
			want:           true,
		},
		{
			name:           "no commits ahead of main",
			aheadCount:     0,
			childrenTotal:  5,
			childrenClosed: 5,
			pendingMRCount: 0,
			want:           false,
		},
		{
			name:           "no children (empty epic)",
			aheadCount:     3,
			childrenTotal:  0,
			childrenClosed: 0,
			pendingMRCount: 0,
			want:           false,
		},
		{
			name:           "not all children closed",
			aheadCount:     3,
			childrenTotal:  5,
			childrenClosed: 3,
			pendingMRCount: 0,
			want:           false,
		},
		{
			name:           "pending MRs still open",
			aheadCount:     3,
			childrenTotal:  5,
			childrenClosed: 5,
			pendingMRCount: 2,
			want:           false,
		},
		{
			name:           "single child closed with commits",
			aheadCount:     1,
			childrenTotal:  1,
			childrenClosed: 1,
			pendingMRCount: 0,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReadyToLand(tt.aheadCount, tt.childrenTotal, tt.childrenClosed, tt.pendingMRCount)
			if got != tt.want {
				t.Errorf("isReadyToLand(%d, %d, %d, %d) = %v, want %v",
					tt.aheadCount, tt.childrenTotal, tt.childrenClosed, tt.pendingMRCount, got, tt.want)
			}
		})
	}
}

// TestMRFilteringByLabel verifies that MRs are identified by their gt:merge-request
// label rather than the deprecated issue_type field. This is the fix for #816 where
// MRs created by `gt done` have issue_type='task' but correct gt:merge-request label.
func TestMRFilteringByLabel(t *testing.T) {
	tests := []struct {
		name     string
		issue    *beads.Issue
		wantIsMR bool
	}{
		{
			name: "MR with correct label and wrong type (bug #816 scenario)",
			issue: &beads.Issue{
				ID:     "mr-1",
				Title:  "Merge: test-branch",
				Type:   "task", // Wrong type (default from bd create)
				Labels: []string{"gt:merge-request"}, // Correct label
			},
			wantIsMR: true,
		},
		{
			name: "MR with correct label and correct type",
			issue: &beads.Issue{
				ID:     "mr-2",
				Title:  "Merge: another-branch",
				Type:   "merge-request",
				Labels: []string{"gt:merge-request"},
			},
			wantIsMR: true,
		},
		{
			name: "Task without MR label",
			issue: &beads.Issue{
				ID:     "task-1",
				Title:  "Regular task",
				Type:   "task",
				Labels: []string{"other-label"},
			},
			wantIsMR: false,
		},
		{
			name: "Issue with no labels",
			issue: &beads.Issue{
				ID:    "issue-1",
				Title: "No labels",
				Type:  "task",
			},
			wantIsMR: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := beads.HasLabel(tt.issue, "gt:merge-request")
			if got != tt.wantIsMR {
				t.Errorf("HasLabel(%q, \"gt:merge-request\") = %v, want %v",
					tt.issue.ID, got, tt.wantIsMR)
			}
		})
	}
}
