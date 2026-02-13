package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	toon "github.com/toon-format/toon-go"
)

// --- Realistic data structures mirroring actual gt command output ---

// Mirrors gt polecat list output
type benchPolecat struct {
	Name      string `json:"name" toon:"name"`
	Rig       string `json:"rig" toon:"rig"`
	Branch    string `json:"branch" toon:"branch"`
	Status    string `json:"status" toon:"status"`
	Session   string `json:"session" toon:"session"`
	HookedBead string `json:"hooked_bead" toon:"hooked_bead"`
	CreatedAt string `json:"created_at" toon:"created_at"`
}

// Mirrors gt status output
type benchTownStatus struct {
	Town       string            `json:"town" toon:"town"`
	Branch     string            `json:"branch" toon:"branch"`
	Rigs       []benchRigStatus  `json:"rigs" toon:"rigs"`
	MailCount  int               `json:"mail_count" toon:"mail_count"`
	DoltStatus string            `json:"dolt_status" toon:"dolt_status"`
}

type benchRigStatus struct {
	Name       string `json:"name" toon:"name"`
	Repo       string `json:"repo" toon:"repo"`
	Branch     string `json:"branch" toon:"branch"`
	Polecats   int    `json:"polecats" toon:"polecats"`
	OpenIssues int    `json:"open_issues" toon:"open_issues"`
	MRsPending int    `json:"mrs_pending" toon:"mrs_pending"`
}

// Mirrors bd list output (issue list)
type benchIssue struct {
	ID          string   `json:"id" toon:"id"`
	Title       string   `json:"title" toon:"title"`
	Status      string   `json:"status" toon:"status"`
	Type        string   `json:"type" toon:"type"`
	Priority    int      `json:"priority" toon:"priority"`
	Assignee    string   `json:"assignee" toon:"assignee"`
	CreatedAt   string   `json:"created_at" toon:"created_at"`
	UpdatedAt   string   `json:"updated_at" toon:"updated_at"`
	Labels      []string `json:"labels" toon:"labels"`
	Description string   `json:"description" toon:"description"`
}

// Mirrors gt mail inbox output
type benchMailMessage struct {
	ID      string `json:"id" toon:"id"`
	From    string `json:"from" toon:"from"`
	To      string `json:"to" toon:"to"`
	Subject string `json:"subject" toon:"subject"`
	Body    string `json:"body" toon:"body"`
	SentAt  string `json:"sent_at" toon:"sent_at"`
	Read    bool   `json:"read" toon:"read"`
}

// --- Data generators ---

func makePolecat(i int) benchPolecat {
	statuses := []string{"running", "idle", "completed", "failed"}
	return benchPolecat{
		Name:       fmt.Sprintf("polecat-%d", i),
		Rig:        "sfgastown",
		Branch:     fmt.Sprintf("feature/st-abc%d", i),
		Status:     statuses[i%len(statuses)],
		Session:    fmt.Sprintf("sfgastown-polecat-%d", i),
		HookedBead: fmt.Sprintf("st-abc%d", i),
		CreatedAt:  "2026-02-13T10:30:00Z",
	}
}

func makePolecats(n int) []benchPolecat {
	out := make([]benchPolecat, n)
	for i := range out {
		out[i] = makePolecat(i)
	}
	return out
}

func makeTownStatus() benchTownStatus {
	return benchTownStatus{
		Town:   "/home/ubuntu/gt",
		Branch: "main",
		Rigs: []benchRigStatus{
			{Name: "sfgastown", Repo: "sfncore/sf-gastown", Branch: "main", Polecats: 3, OpenIssues: 12, MRsPending: 2},
			{Name: "frankencord", Repo: "sfncore/openclaw", Branch: "main", Polecats: 1, OpenIssues: 5, MRsPending: 0},
			{Name: "frankentui", Repo: "Dicklesworthstone/frankentui", Branch: "main", Polecats: 0, OpenIssues: 8, MRsPending: 1},
		},
		MailCount:  7,
		DoltStatus: "running",
	}
}

func makeIssue(i int) benchIssue {
	statuses := []string{"open", "in_progress", "closed", "blocked"}
	types := []string{"task", "bug", "feature", "epic"}
	return benchIssue{
		ID:        fmt.Sprintf("st-%04d", i),
		Title:     fmt.Sprintf("Implement feature %d: add support for advanced configuration options", i),
		Status:    statuses[i%len(statuses)],
		Type:      types[i%len(types)],
		Priority:  i % 5,
		Assignee:  "polecat-1",
		CreatedAt: "2026-02-10T08:00:00Z",
		UpdatedAt: "2026-02-13T12:00:00Z",
		Labels:    []string{"enhancement", "v2"},
		Description: fmt.Sprintf("This task implements feature %d. It requires changes to the config parser, "+
			"the validation layer, and the CLI flags. Acceptance criteria: all existing tests pass, "+
			"new tests cover the added functionality, documentation is updated.", i),
	}
}

func makeIssues(n int) []benchIssue {
	out := make([]benchIssue, n)
	for i := range out {
		out[i] = makeIssue(i)
	}
	return out
}

func makeMailMessages(n int) []benchMailMessage {
	out := make([]benchMailMessage, n)
	for i := range out {
		out[i] = benchMailMessage{
			ID:      fmt.Sprintf("hq-mail-%04d", i),
			From:    "sfgastown/crew/polecat-1",
			To:      "mayor/",
			Subject: fmt.Sprintf("MR Ready: st-%04d implementation complete", i),
			Body: fmt.Sprintf("The implementation for st-%04d is complete. Branch feature/st-%04d "+
				"has 3 commits. All tests pass. Please review and merge.\n\n"+
				"Changes:\n- Modified internal/cmd/config.go\n- Added internal/cmd/config_test.go\n"+
				"- Updated docs/config.md", i, i),
			SentAt: "2026-02-13T11:30:00Z",
			Read:   i%3 == 0,
		}
	}
	return out
}

// --- Helper: discard stdout during benchmarks ---

func discardStdout(b *testing.B) func() {
	b.Helper()
	old := os.Stdout
	os.Stdout = devNull(b)
	return func() { os.Stdout = old }
}

func devNull(b *testing.B) *os.File {
	b.Helper()
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	return f
}

// --- Size comparison test (not a benchmark, but shows the savings) ---

func TestTOONSizeSavings(t *testing.T) {
	cases := []struct {
		name string
		data any
	}{
		{"polecats_5", makePolecats(5)},
		{"polecats_20", makePolecats(20)},
		{"town_status", makeTownStatus()},
		{"issues_10", makeIssues(10)},
		{"issues_50", makeIssues(50)},
		{"mail_10", makeMailMessages(10)},
		{"mail_50", makeMailMessages(50)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBytes, err := json.MarshalIndent(tc.data, "", "  ")
			if err != nil {
				t.Fatalf("json marshal: %v", err)
			}
			toonBytes, err := toon.Marshal(tc.data)
			if err != nil {
				t.Fatalf("toon marshal: %v", err)
			}

			jsonSize := len(jsonBytes)
			toonSize := len(toonBytes)
			saving := float64(jsonSize-toonSize) / float64(jsonSize) * 100

			t.Logf("JSON: %5d bytes | TOON: %5d bytes | saving: %.1f%%", jsonSize, toonSize, saving)

			if toonSize >= jsonSize {
				t.Errorf("TOON (%d) should be smaller than JSON (%d) for tabular data", toonSize, jsonSize)
			}
		})
	}
}

// --- Marshaling benchmarks (raw speed, no I/O) ---

func BenchmarkMarshal_Polecats5(b *testing.B) {
	data := makePolecats(5)
	b.Run("json", func(b *testing.B) {
		for b.Loop() {
			json.MarshalIndent(data, "", "  ")
		}
	})
	b.Run("toon", func(b *testing.B) {
		for b.Loop() {
			toon.Marshal(data)
		}
	})
}

func BenchmarkMarshal_Polecats20(b *testing.B) {
	data := makePolecats(20)
	b.Run("json", func(b *testing.B) {
		for b.Loop() {
			json.MarshalIndent(data, "", "  ")
		}
	})
	b.Run("toon", func(b *testing.B) {
		for b.Loop() {
			toon.Marshal(data)
		}
	})
}

func BenchmarkMarshal_TownStatus(b *testing.B) {
	data := makeTownStatus()
	b.Run("json", func(b *testing.B) {
		for b.Loop() {
			json.MarshalIndent(data, "", "  ")
		}
	})
	b.Run("toon", func(b *testing.B) {
		for b.Loop() {
			toon.Marshal(data)
		}
	})
}

func BenchmarkMarshal_Issues10(b *testing.B) {
	data := makeIssues(10)
	b.Run("json", func(b *testing.B) {
		for b.Loop() {
			json.MarshalIndent(data, "", "  ")
		}
	})
	b.Run("toon", func(b *testing.B) {
		for b.Loop() {
			toon.Marshal(data)
		}
	})
}

func BenchmarkMarshal_Issues50(b *testing.B) {
	data := makeIssues(50)
	b.Run("json", func(b *testing.B) {
		for b.Loop() {
			json.MarshalIndent(data, "", "  ")
		}
	})
	b.Run("toon", func(b *testing.B) {
		for b.Loop() {
			toon.Marshal(data)
		}
	})
}

func BenchmarkMarshal_Mail10(b *testing.B) {
	data := makeMailMessages(10)
	b.Run("json", func(b *testing.B) {
		for b.Loop() {
			json.MarshalIndent(data, "", "  ")
		}
	})
	b.Run("toon", func(b *testing.B) {
		for b.Loop() {
			toon.Marshal(data)
		}
	})
}

func BenchmarkMarshal_Mail50(b *testing.B) {
	data := makeMailMessages(50)
	b.Run("json", func(b *testing.B) {
		for b.Loop() {
			json.MarshalIndent(data, "", "  ")
		}
	})
	b.Run("toon", func(b *testing.B) {
		for b.Loop() {
			toon.Marshal(data)
		}
	})
}

// --- End-to-end Print benchmarks (marshal + write, the full path) ---

func BenchmarkPrint_Polecats5(b *testing.B) {
	data := makePolecats(5)
	b.Run("json", func(b *testing.B) {
		restore := discardStdout(b)
		defer restore()
		for b.Loop() {
			PrintJSON(data)
		}
	})
	b.Run("toon", func(b *testing.B) {
		restore := discardStdout(b)
		defer restore()
		for b.Loop() {
			PrintTOON(data)
		}
	})
}

func BenchmarkPrint_Issues50(b *testing.B) {
	data := makeIssues(50)
	b.Run("json", func(b *testing.B) {
		restore := discardStdout(b)
		defer restore()
		for b.Loop() {
			PrintJSON(data)
		}
	})
	b.Run("toon", func(b *testing.B) {
		restore := discardStdout(b)
		defer restore()
		for b.Loop() {
			PrintTOON(data)
		}
	})
}

func BenchmarkPrint_Mail50(b *testing.B) {
	data := makeMailMessages(50)
	b.Run("json", func(b *testing.B) {
		restore := discardStdout(b)
		defer restore()
		for b.Loop() {
			PrintJSON(data)
		}
	})
	b.Run("toon", func(b *testing.B) {
		restore := discardStdout(b)
		defer restore()
		for b.Loop() {
			PrintTOON(data)
		}
	})
}

// --- Output size benchmarks (report bytes written) ---

func BenchmarkSize_Issues50(b *testing.B) {
	data := makeIssues(50)

	b.Run("json", func(b *testing.B) {
		var totalBytes int64
		for b.Loop() {
			out, _ := json.MarshalIndent(data, "", "  ")
			totalBytes += int64(len(out))
		}
		b.ReportMetric(float64(totalBytes)/float64(b.N), "bytes/op")
	})

	b.Run("toon", func(b *testing.B) {
		var totalBytes int64
		for b.Loop() {
			out, _ := toon.Marshal(data)
			totalBytes += int64(len(out))
		}
		b.ReportMetric(float64(totalBytes)/float64(b.N), "bytes/op")
	})
}

// --- JSON round-trip benchmark (simulates bd's path) ---

func BenchmarkRoundTrip_Issues10(b *testing.B) {
	data := makeIssues(10)

	b.Run("direct_json", func(b *testing.B) {
		for b.Loop() {
			json.MarshalIndent(data, "", "  ")
		}
	})

	b.Run("roundtrip_toon", func(b *testing.B) {
		// This is bd's path: json.Marshal -> json.Unmarshal -> toon.Marshal
		for b.Loop() {
			jsonBytes, _ := json.Marshal(data)
			var generic any
			json.Unmarshal(jsonBytes, &generic)
			toon.Marshal(generic)
		}
	})
}

// --- Comparison helper: prints a summary table when run with -v ---

func TestBenchmarkSummary(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping summary in short mode")
	}

	cases := []struct {
		name string
		data any
	}{
		{"5 polecats (gt polecat list)", makePolecats(5)},
		{"20 polecats", makePolecats(20)},
		{"town status (gt status)", makeTownStatus()},
		{"10 issues (bd list)", makeIssues(10)},
		{"50 issues", makeIssues(50)},
		{"10 mail messages (gt mail inbox)", makeMailMessages(10)},
		{"50 mail messages", makeMailMessages(50)},
	}

	t.Log("")
	t.Log("┌──────────────────────────────────┬───────────┬───────────┬──────────┐")
	t.Log("│ Dataset                          │ JSON (B)  │ TOON (B)  │ Saving   │")
	t.Log("├──────────────────────────────────┼───────────┼───────────┼──────────┤")

	for _, tc := range cases {
		jsonBytes, _ := json.MarshalIndent(tc.data, "", "  ")
		toonBytes, _ := toon.Marshal(tc.data)
		saving := float64(len(jsonBytes)-len(toonBytes)) / float64(len(jsonBytes)) * 100
		t.Logf("│ %-32s │ %7d   │ %7d   │ %5.1f%%   │", tc.name, len(jsonBytes), len(toonBytes), saving)
	}

	t.Log("└──────────────────────────────────┴───────────┴───────────┴──────────┘")

	// Also show the bd round-trip overhead
	issues10 := makeIssues(10)
	directJSON, _ := json.MarshalIndent(issues10, "", "  ")
	jsonBytes, _ := json.Marshal(issues10)
	var generic any
	json.Unmarshal(jsonBytes, &generic)
	roundtripTOON, _ := toon.Marshal(generic)

	t.Log("")
	t.Logf("bd round-trip (10 issues): JSON %d B → TOON %d B (%.1f%% saving)",
		len(directJSON), len(roundtripTOON),
		float64(len(directJSON)-len(roundtripTOON))/float64(len(directJSON))*100)
}

// sink prevents compiler from optimizing away results
var sink []byte
var sinkErr error

func init() {
	// Prevent "unused" warnings
	_ = sink
	_ = sinkErr
	_ = io.Discard
}
