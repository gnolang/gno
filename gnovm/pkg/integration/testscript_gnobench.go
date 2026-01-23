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
	p.Cmds["jsonbench"] = CmdJSONBench
	p.Cmds["cmpbench"] = CmdCmpBench

	return nil
}

// ---- Exported Bench State and Commands for Reuse

const defaultProfileFile = "profile.json"

// BenchState tracks profiling state during a testscript test.
type BenchState struct {
	Running    bool
	OutputFile string
	Files      []string // All generated files for golden comparison
}

// RegisterBenchCommands adds benchops-related commands to testscript params.
// It registers "bench", "jsonbench", and "cmpbench" commands.
// The stateKey is used to store/retrieve BenchState from testscript.Env.Values.
func RegisterBenchCommands(p *testscript.Params, stateKey any) {
	if p.Cmds == nil {
		p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	}
	p.Cmds["bench"] = makeCmdBench(stateKey)
	p.Cmds["jsonbench"] = CmdJSONBench
	p.Cmds["cmpbench"] = CmdCmpBench
}

// makeCmdBench creates the "bench" command handler.
// Usage: bench start [filename] | bench stop
func makeCmdBench(stateKey any) func(*testscript.TestScript, bool, []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if neg {
			ts.Fatalf("bench command does not support negation")
		}

		if len(args) == 0 {
			ts.Fatalf("usage: bench start [file] | bench stop")
		}

		// Retrieve state from testscript values
		state, ok := ts.Value(stateKey).(*BenchState)
		if !ok {
			ts.Fatalf("bench: state not found (setup may be missing)")
		}

		CmdBenchWithState(ts, state, args)
	}
}

// AutoStopBench stops profiling if still running and writes results.
// Call this at end of test to ensure bench data is captured.
func AutoStopBench(ts *testscript.TestScript, state *BenchState) {
	if !state.Running {
		return
	}

	results := benchops.Stop()
	state.Running = false

	path := ts.MkAbs(state.OutputFile)
	if err := writeBenchResults(path, results); err != nil {
		ts.Logf("bench: auto-stop failed: %v", err)
		return
	}

	state.Files = append(state.Files, state.OutputFile)
	ts.Logf("bench: auto-stopped profiling, wrote %s", state.OutputFile)
}

// writeBenchResults writes benchmark results to a file with proper error handling.
func writeBenchResults(path string, results *benchops.Results) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close output file: %w", cerr)
		}
	}()

	if err := results.WriteJSON(f); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}
	return nil
}

// FormatBenchOutput formats bench profile data for golden comparison.
// Only includes deterministic fields (Count, Gas).
func FormatBenchOutput(data []byte) (string, error) {
	events, err := parseBenchData(data)
	if err != nil {
		return "", err
	}

	if len(events) == 0 {
		return "<no bench events>\n", nil
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
	return trimLines(out.String()), nil
}

// parseBenchData parses bench profile data, auto-detecting the format.
// Supports both formats:
//   - Direct benchops.Results JSON (from "gno run --bench-profile")
//   - JSONL with benchEvent wrapper (from "gno test --bench-profile")
func parseBenchData(data []byte) ([]benchEvent, error) {
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, nil
	}

	// Try parsing as direct profile first (benchops.Results)
	var profile benchops.Results
	if err := json.Unmarshal([]byte(content), &profile); err == nil {
		if profile.OpStats != nil {
			return []benchEvent{{Profile: &profile}}, nil
		}
	}

	// Parse as JSONL (benchEvent with Profile)
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

// ---- Bench JSON Parsing (uses benchops types)

// benchEvent is the format from "gno test --bench-profile" (JSONL with wrapper).
// Uses benchops.Results for the profile data to maintain single source of truth.
type benchEvent struct {
	Package string            `json:"Package"`
	Test    string            `json:"Test"`
	Profile *benchops.Results `json:"Profile,omitempty"`
}

// ---- Command Implementation

// CmdJSONBench is the exported jsonbench testscript command.
// Usage: jsonbench <file>
//
// Parses bench profile output and displays deterministic fields.
// Supports both formats:
//   - "gno test --bench-profile": JSONL with BenchEvent wrapper
//   - "gno run --bench-profile": direct JSON with OpStats
//
// With negation (! jsonbench), expects no valid events to be found.
func CmdJSONBench(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 1 {
		ts.Fatalf("usage: jsonbench <file>")
	}

	// File read errors are always fatal, regardless of negation
	events, err := parseBenchFile(ts.MkAbs(args[0]))
	if err != nil {
		ts.Fatalf("jsonbench: %v", err)
	}

	hasEvents := len(events) > 0

	if neg {
		// With negation, expect no valid events
		if hasEvents {
			ts.Fatalf("jsonbench: expected no events but found %d", len(events))
		}
		return
	}

	// Without negation, expect valid events
	if !hasEvents {
		ts.Fatalf("jsonbench: no events found")
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

// CmdCmpBench compares two bench profile files by their deterministic fields.
// Usage: cmpbench <file1> <file2>
//
// Compares only Count and Gas values, ignoring timing data.
// Fails if the deterministic fields differ.
// With negation (! cmpbench), expects files to differ.
func CmdCmpBench(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) != 2 {
		ts.Fatalf("usage: cmpbench <file1> <file2>")
	}

	// File read errors are always fatal, regardless of negation
	formatted1, err := formatBenchFile(ts.MkAbs(args[0]))
	if err != nil {
		ts.Fatalf("cmpbench: %s: %v", args[0], err)
	}

	formatted2, err := formatBenchFile(ts.MkAbs(args[1]))
	if err != nil {
		ts.Fatalf("cmpbench: %s: %v", args[1], err)
	}

	differ := formatted1 != formatted2

	if neg {
		// With negation, expect files to differ
		if !differ {
			ts.Fatalf("cmpbench: %s and %s are identical (expected difference)", args[0], args[1])
		}
		return
	}

	// Without negation, expect files to be identical
	if differ {
		ts.Fatalf("cmpbench: %s and %s differ:\n--- %s\n%s\n--- %s\n%s",
			args[0], args[1], args[0], formatted1, args[1], formatted2)
	}
}

// formatBenchFile reads and formats a bench profile file for comparison.
func formatBenchFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return FormatBenchOutput(data)
}

// CmdBenchWithState implements the bench command with an external state.
// This allows callers to manage state separately from testscript env.Values.
func CmdBenchWithState(ts *testscript.TestScript, state *BenchState, args []string) {
	switch args[0] {
	case "start":
		if state.Running {
			ts.Fatalf("bench: profiler already running (missing bench stop?)")
		}

		state.OutputFile = defaultProfileFile
		if len(args) > 1 {
			state.OutputFile = args[1]
		}

		benchops.Start()
		state.Running = true
		ts.Logf("bench: started profiling → %s", state.OutputFile)

	case "stop":
		if !state.Running {
			ts.Fatalf("bench: profiler not running (missing bench start?)")
		}

		results := benchops.Stop()
		state.Running = false

		// Write JSON to file in the work directory
		path := ts.MkAbs(state.OutputFile)
		if err := writeBenchResults(path, results); err != nil {
			ts.Fatalf("bench: %v", err)
		}

		state.Files = append(state.Files, state.OutputFile)
		ts.Logf("bench: stopped profiling, wrote %s", state.OutputFile)

	default:
		ts.Fatalf("bench: unknown subcommand %q (use 'start' or 'stop')", args[0])
	}
}

// parseBenchFile parses a bench profile file, auto-detecting the format.
func parseBenchFile(filename string) ([]benchEvent, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseBenchData(data)
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
