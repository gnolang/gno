//go:build gnobench

package integration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/rogpeppe/go-internal/testscript"
)

// SetupGnoBench prepares the given testscript environment for tests that utilize
// the gno command built with -tags gnobench. This enables benchops profiling.
func SetupGnoBench(p *testscript.Params, homeDir, buildDir string) error {
	gnoroot := gnoenv.RootDir()

	gnoBin, err := buildGnoBinary(gnoroot, buildDir, GnoBuildOptions{
		BinaryName: "gno_bench",
		BuildTags:  "gnobench",
	})
	if err != nil {
		return err
	}

	setupGnoCommand(p, gnoBin, gnoroot, homeDir)

	if p.Cmds == nil {
		p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	}
	p.Cmds["jsonbench"] = cmdJSONBench

	return nil
}

// ---- Bench JSON Parsing (uses benchops types)

// benchEvent is the format from "gno test --bench-profile" (JSONL with wrapper).
// Uses benchops.Results for the profile data to maintain single source of truth.
type benchEvent struct {
	Package string           `json:"Package"`
	Test    string           `json:"Test"`
	Profile *benchops.Results `json:"Profile,omitempty"`
}

// ---- Command Implementation

// cmdJSONBench implements the jsonbench testscript command.
// Usage: jsonbench <file>
//
// Parses bench profile output and displays deterministic fields.
// Supports both formats:
//   - "gno test --bench-profile": JSONL with BenchEvent wrapper
//   - "gno run --bench-profile": direct JSON with OpStats
func cmdJSONBench(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 1 {
		ts.Fatalf("usage: jsonbench <file>")
	}

	events, err := parseBenchFile(ts.MkAbs(args[0]))
	if err != nil {
		if neg {
			return
		}
		ts.Fatalf("jsonbench: %v", err)
	}

	if len(events) == 0 {
		if neg {
			return
		}
		ts.Fatalf("jsonbench: no events found")
	}

	if neg {
		ts.Fatalf("jsonbench: expected failure but succeeded")
	}

	var out strings.Builder
	tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)

	for i, event := range events {
		if i > 0 {
			fmt.Fprintln(tw)
		}
		writeEvent(tw, &event)
	}

	tw.Flush()
	fmt.Fprint(ts.Stdout(), trimLines(out.String()))
}

// parseBenchFile parses a bench profile file, auto-detecting the format.
func parseBenchFile(filename string) ([]benchEvent, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, nil
	}

	// Try parsing as direct profile first (gno run format - benchops.Results)
	var profile benchops.Results
	if err := json.Unmarshal([]byte(content), &profile); err == nil {
		if profile.OpStats != nil {
			return []benchEvent{{Profile: &profile}}, nil
		}
	}

	// Parse as JSONL (gno test format - benchEvent with Profile)
	var events []benchEvent
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event benchEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		events = append(events, event)
	}

	return events, scanner.Err()
}

func writeEvent(tw *tabwriter.Writer, event *benchEvent) {
	// Only print Package/Test if present (gno test format)
	if event.Package != "" {
		fmt.Fprintf(tw, "Package:\t%s\n", event.Package)
	}
	if event.Test != "" {
		fmt.Fprintf(tw, "Test:\t%s\n", event.Test)
	}

	// Always print OpStats (deterministic fields only: Count, Gas)
	fmt.Fprintln(tw, "OpStats:")
	if event.Profile == nil || len(event.Profile.OpStats) == 0 {
		fmt.Fprintln(tw, "  <none>")
		return
	}

	fmt.Fprintf(tw, "  Name\tCount\tGas\n")
	keys := make([]string, 0, len(event.Profile.OpStats))
	for k := range event.Profile.OpStats {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		stat := event.Profile.OpStats[name]
		// Only output deterministic fields from benchops.OpStatJSON
		fmt.Fprintf(tw, "  %s\t%d\t%d\n", name, stat.Count, stat.Gas)
	}
}

func trimLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}
