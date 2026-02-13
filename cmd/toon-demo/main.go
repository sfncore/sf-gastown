// Quick demo of TOON vs JSON output for gt commands.
// Usage: go run ./cmd/toon-demo
package main

import (
	"fmt"
	"os"

	"github.com/steveyegge/gastown/internal/output"
)

type Agent struct {
	Name     string `json:"name" toon:"name"`
	Role     string `json:"role" toon:"role"`
	Runtime  string `json:"runtime" toon:"runtime"`
	Rig      string `json:"rig,omitempty" toon:"rig"`
	Attached bool   `json:"attached" toon:"attached"`
}

type TownStatus struct {
	Name  string  `json:"name" toon:"name"`
	Rigs  []Rig   `json:"rigs" toon:"rigs"`
	Agents []Agent `json:"agents" toon:"agents"`
}

type Rig struct {
	Name     string `json:"name" toon:"name"`
	Polecats int    `json:"polecats" toon:"polecats"`
	Crew     int    `json:"crew" toon:"crew"`
}

func main() {
	status := TownStatus{
		Name: "gt",
		Rigs: []Rig{
			{Name: "sfgastown", Polecats: 2, Crew: 2},
			{Name: "frankencord", Polecats: 0, Crew: 1},
			{Name: "frankentui", Polecats: 0, Crew: 2},
			{Name: "tmux_adapter", Polecats: 0, Crew: 0},
		},
		Agents: []Agent{
			{Name: "hq-mayor", Role: "mayor", Runtime: "claude", Attached: true},
			{Name: "hq-deacon", Role: "deacon", Runtime: "claude", Attached: false},
			{Name: "gt-sfgastown-polecat-fox", Role: "polecat", Runtime: "opencode", Rig: "sfgastown", Attached: false},
			{Name: "gt-sfgastown-polecat-lynx", Role: "polecat", Runtime: "opencode", Rig: "sfgastown", Attached: false},
			{Name: "gt-sfgastown-witness", Role: "witness", Runtime: "opencode", Rig: "sfgastown", Attached: false},
		},
	}

	fmt.Println("=== JSON ===")
	output.PrintJSON(status)

	fmt.Println("\n=== TOON ===")
	output.PrintTOON(status)

	// Test env var
	fmt.Println("\n=== ResolveFormat ===")
	fmt.Printf("Default: %s\n", output.ResolveFormat(""))
	os.Setenv("GT_OUTPUT_FORMAT", "toon")
	fmt.Printf("With GT_OUTPUT_FORMAT=toon: %s\n", output.ResolveFormat(""))
	fmt.Printf("Flag override: %s\n", output.ResolveFormat("json"))
}
