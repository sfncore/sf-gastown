package beads

import (
	"testing"
)

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
			name:        "field present",
			description: "integration_branch: integration/gt-epic",
			want:        "integration/gt-epic",
		},
		{
			name:        "field with surrounding text",
			description: "Some description\nintegration_branch: feature/my-branch\nMore text",
			want:        "feature/my-branch",
		},
		{
			name:        "case insensitive",
			description: "INTEGRATION_BRANCH: integration/GT-123",
			want:        "integration/GT-123",
		},
		{
			name:        "field not present",
			description: "Some description\nbase_branch: develop\n",
			want:        "",
		},
		{
			name:        "field with extra whitespace",
			description: "  integration_branch:   integration/spaced  ",
			want:        "integration/spaced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIntegrationBranchField(tt.description)
			if got != tt.want {
				t.Errorf("GetIntegrationBranchField() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetBaseBranchField(t *testing.T) {
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
			name:        "field present",
			description: "base_branch: develop",
			want:        "develop",
		},
		{
			name:        "alongside integration_branch",
			description: "integration_branch: integration/gt-epic\nbase_branch: release/v2",
			want:        "release/v2",
		},
		{
			name:        "field not present",
			description: "integration_branch: integration/gt-epic",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBaseBranchField(tt.description)
			if got != tt.want {
				t.Errorf("GetBaseBranchField() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddIntegrationBranchField(t *testing.T) {
	tests := []struct {
		name        string
		description string
		branchName  string
		want        string
	}{
		{
			name:        "empty description",
			description: "",
			branchName:  "integration/gt-epic",
			want:        "integration_branch: integration/gt-epic",
		},
		{
			name:        "add to existing",
			description: "Some description",
			branchName:  "integration/gt-epic",
			want:        "integration_branch: integration/gt-epic\nSome description",
		},
		{
			name:        "replace existing",
			description: "integration_branch: old-branch\nSome description",
			branchName:  "integration/new-branch",
			want:        "integration_branch: integration/new-branch\nSome description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddIntegrationBranchField(tt.description, tt.branchName)
			if got != tt.want {
				t.Errorf("AddIntegrationBranchField() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddBaseBranchField(t *testing.T) {
	tests := []struct {
		name        string
		description string
		baseBranch  string
		want        string
	}{
		{
			name:        "empty description",
			description: "",
			baseBranch:  "develop",
			want:        "base_branch: develop",
		},
		{
			name:        "add to existing",
			description: "integration_branch: integration/gt-epic",
			baseBranch:  "develop",
			want:        "base_branch: develop\nintegration_branch: integration/gt-epic",
		},
		{
			name:        "replace existing",
			description: "base_branch: old\nSome text",
			baseBranch:  "release/v2",
			want:        "base_branch: release/v2\nSome text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddBaseBranchField(tt.description, tt.baseBranch)
			if got != tt.want {
				t.Errorf("AddBaseBranchField() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildIntegrationBranchName(t *testing.T) {
	tests := []struct {
		name     string
		template string
		epicID   string
		want     string
	}{
		{
			name:     "default template",
			template: "",
			epicID:   "gt-epic",
			want:     "integration/gt-epic",
		},
		{
			name:     "custom template with epic",
			template: "feature/{epic}",
			epicID:   "RA-123",
			want:     "feature/RA-123",
		},
		{
			name:     "custom template with prefix",
			template: "{prefix}/integration/{epic}",
			epicID:   "PROJ-456",
			want:     "PROJ/integration/PROJ-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildIntegrationBranchName(tt.template, tt.epicID)
			if got != tt.want {
				t.Errorf("BuildIntegrationBranchName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractEpicPrefix(t *testing.T) {
	tests := []struct {
		epicID string
		want   string
	}{
		{"RA-123", "RA"},
		{"PROJ-456", "PROJ"},
		{"abc", "abc"},
		{"a-b-c", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.epicID, func(t *testing.T) {
			got := ExtractEpicPrefix(tt.epicID)
			if got != tt.want {
				t.Errorf("ExtractEpicPrefix(%q) = %q, want %q", tt.epicID, got, tt.want)
			}
		})
	}
}

// mockBranchChecker implements BranchChecker for testing.
type mockBranchChecker struct {
	localBranches  map[string]bool
	remoteBranches map[string]bool // key: "remote/branch"
}

func (m *mockBranchChecker) BranchExists(name string) (bool, error) {
	return m.localBranches[name], nil
}

func (m *mockBranchChecker) RemoteBranchExists(remote, name string) (bool, error) {
	key := remote + "/" + name
	return m.remoteBranches[key], nil
}

func TestDetectIntegrationBranch(t *testing.T) {
	// Create a mock beads that returns predefined issues.
	// We can't easily mock Beads since it shells out to bd,
	// so we test the logic through the exported functions instead.
	// The integration of DetectIntegrationBranch with real Beads
	// is covered by the manual verification steps.

	// Test the metadata extraction logic that DetectIntegrationBranch relies on
	t.Run("metadata extraction for detect logic", func(t *testing.T) {
		// Simulate an epic with integration_branch metadata
		desc := "integration_branch: feature/custom-branch"
		branch := GetIntegrationBranchField(desc)
		if branch != "feature/custom-branch" {
			t.Errorf("expected 'feature/custom-branch', got %q", branch)
		}

		// Simulate an epic without metadata - fallback to naming convention
		desc2 := "Some epic description"
		branch2 := GetIntegrationBranchField(desc2)
		if branch2 != "" {
			t.Errorf("expected empty string, got %q", branch2)
		}
		// In this case, DetectIntegrationBranch would call BuildIntegrationBranchName
		fallback := BuildIntegrationBranchName("", "gt-epic")
		if fallback != "integration/gt-epic" {
			t.Errorf("expected 'integration/gt-epic', got %q", fallback)
		}
	})

	// Test BranchChecker mock
	t.Run("branch checker mock", func(t *testing.T) {
		checker := &mockBranchChecker{
			localBranches: map[string]bool{
				"integration/gt-epic": true,
			},
			remoteBranches: map[string]bool{
				"origin/integration/gt-other": true,
			},
		}

		exists, err := checker.BranchExists("integration/gt-epic")
		if err != nil || !exists {
			t.Errorf("BranchExists should return true for local branch")
		}

		exists, err = checker.BranchExists("integration/gt-missing")
		if err != nil || exists {
			t.Errorf("BranchExists should return false for missing branch")
		}

		exists, err = checker.RemoteBranchExists("origin", "integration/gt-other")
		if err != nil || !exists {
			t.Errorf("RemoteBranchExists should return true for remote branch")
		}
	})
}
