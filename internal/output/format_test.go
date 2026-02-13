package output

import (
	"os"
	"testing"
)

func TestResolveFormat(t *testing.T) {
	// Clean env for each test
	orig := os.Getenv("GT_OUTPUT_FORMAT")
	defer os.Setenv("GT_OUTPUT_FORMAT", orig)

	t.Run("default is json", func(t *testing.T) {
		os.Unsetenv("GT_OUTPUT_FORMAT")
		if got := ResolveFormat(""); got != FormatJSON {
			t.Errorf("ResolveFormat(\"\") = %q, want %q", got, FormatJSON)
		}
	})

	t.Run("explicit flag wins", func(t *testing.T) {
		os.Setenv("GT_OUTPUT_FORMAT", "json")
		if got := ResolveFormat("toon"); got != FormatTOON {
			t.Errorf("ResolveFormat(\"toon\") = %q, want %q", got, FormatTOON)
		}
	})

	t.Run("env var when no flag", func(t *testing.T) {
		os.Setenv("GT_OUTPUT_FORMAT", "toon")
		if got := ResolveFormat(""); got != FormatTOON {
			t.Errorf("ResolveFormat(\"\") with GT_OUTPUT_FORMAT=toon = %q, want %q", got, FormatTOON)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		os.Unsetenv("GT_OUTPUT_FORMAT")
		if got := ResolveFormat("TOON"); got != FormatTOON {
			t.Errorf("ResolveFormat(\"TOON\") = %q, want %q", got, FormatTOON)
		}
	})
}

func TestPrintFormatted(t *testing.T) {
	type testAgent struct {
		Name    string `json:"name" toon:"name"`
		Role    string `json:"role" toon:"role"`
		Runtime string `json:"runtime" toon:"runtime"`
	}

	agents := []testAgent{
		{Name: "hq-mayor", Role: "mayor", Runtime: "claude"},
		{Name: "hq-deacon", Role: "deacon", Runtime: "claude"},
	}

	// Just verify no errors — actual output goes to stdout
	t.Run("json format", func(t *testing.T) {
		if err := PrintFormatted(agents, FormatJSON); err != nil {
			t.Errorf("PrintFormatted JSON error: %v", err)
		}
	})

	t.Run("toon format", func(t *testing.T) {
		if err := PrintFormatted(agents, FormatTOON); err != nil {
			t.Errorf("PrintFormatted TOON error: %v", err)
		}
	})
}
