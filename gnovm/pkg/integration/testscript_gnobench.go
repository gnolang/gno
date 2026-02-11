//go:build gnobench

package integration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/rogpeppe/go-internal/testscript"
)

// benchMutex serializes access to the global benchops profiler across parallel tests.
var benchMutex sync.Mutex

// benchStateKey is the key used to store BenchState in testscript env.Values.
type benchStateKey struct{}

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

	// Wrap Setup to initialize BenchState for each test
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}
		env.Values[benchStateKey{}] = &BenchState{}
		return nil
	}

	if p.Cmds == nil {
		p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	}
	p.Cmds["bench"] = makeCmdBench(benchStateKey{})
	p.Cmds["jsonbench"] = CmdJSONBench
	p.Cmds["cmpbench"] = CmdCmpBench

	return nil
}

// ---- Exported Bench State and Commands for Reuse

// BenchState tracks profiling state during a testscript test.
type BenchState struct {
	Running    bool                  // Whether profiling is currently active
	OutputFile string                // Target file for profile output
	Sections   benchops.SectionFlags // Sections to include in golden output
	Files      []string              // All generated profile files for comparison
	mutexHeld  bool                  // Whether this test holds the benchMutex
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
// Usage: bench start <filename.json> | bench stop
//
// The filename MUST end with .json or .jsonl extension.
// Example: bench start profile.json
func makeCmdBench(stateKey any) func(*testscript.TestScript, bool, []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if neg {
			ts.Fatalf("bench command does not support negation")
		}

		if len(args) == 0 {
			ts.Fatalf("usage: bench start <filename.json> | bench stop")
		}

		// Retrieve state from testscript values
		state, ok := ts.Value(stateKey).(*BenchState)
		if !ok {
			ts.Fatalf("bench: state not found (setup may be missing)")
		}

		CmdBenchWithState(ts, state, args)
	}
}

// AutoStopBench ensures bench data is captured at end of test.
func AutoStopBench(ts *testscript.TestScript, state *BenchState) {
	if !state.Running {
		return
	}

	results := benchops.Stop()
	state.Running = false

	if state.mutexHeld {
		benchMutex.Unlock()
		state.mutexHeld = false
	}

	path := ts.MkAbs(state.OutputFile)
	if err := writeBenchResults(path, results); err != nil {
		ts.Logf("bench: auto-stop failed: %v", err)
		return
	}

	state.Files = append(state.Files, state.OutputFile)
	ts.Logf("bench: auto-stopped profiling, wrote %s", state.OutputFile)
}

// writeBenchResults writes benchmark results as JSON to a file.
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

	if werr := results.WriteJSON(f); werr != nil {
		return fmt.Errorf("failed to write JSON: %w", werr)
	}
	return nil
}

// FormatBenchOutput formats bench profile data for golden comparison.
// Uses WriteGolden for deterministic output (excludes timing, sorts alphabetically).
// The sections parameter controls which sections to include (0 = all).
func FormatBenchOutput(data []byte, sections benchops.SectionFlags) (string, error) {
	events, err := parseBenchData(data)
	if err != nil {
		return "", err
	}

	if len(events) == 0 {
		return "<no bench events>\n", nil
	}

	var out strings.Builder
	for i, event := range events {
		if i > 0 {
			fmt.Fprintln(&out)
		}
		writeEventGolden(&out, &event, sections)
	}

	return trimLines(out.String()), nil
}

// parseBenchData parses bench profile data, auto-detecting the format.
// Supports both formats:
//   - Direct benchops.Results JSON (from "gno run --bench-profile")
//   - JSONL with benchops.BenchEvent wrapper (from "gno test --bench-profile")
func parseBenchData(data []byte) ([]benchops.BenchEvent, error) {
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, nil
	}

	// Try parsing as direct profile first (benchops.Results)
	var profile benchops.Results
	if err := json.Unmarshal([]byte(content), &profile); err == nil {
		if profile.OpStats != nil {
			return []benchops.BenchEvent{{Profile: &profile}}, nil
		}
	}

	// Parse as JSONL (benchops.BenchEvent with Profile)
	var events []benchops.BenchEvent
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event benchops.BenchEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		events = append(events, event)
	}

	return events, scanner.Err()
}

// CmdJSONBench parses and displays a bench profile file in golden format.
// Usage: jsonbench <file.json> [sections]
//
// The file MUST end with .json or .jsonl extension.
// The optional sections parameter is a comma-separated list (e.g., "opcodes,native").
// Outputs deterministic golden format for comparison.
// With negation (! jsonbench), expects no valid events to be found.
func CmdJSONBench(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) < 1 || len(args) > 2 {
		ts.Fatalf("usage: jsonbench <file.json> [sections]")
	}

	filename := args[0]

	// Validate .json/.jsonl extension using centralized function
	if !benchops.IsJSONFormat(filename) {
		ts.Fatalf("jsonbench: file must end with .json or .jsonl, got %q", filename)
	}

	var sections benchops.SectionFlags // default: 0 = all
	if len(args) == 2 {
		var err error
		sections, err = benchops.ParseSectionFlags(args[1])
		if err != nil {
			ts.Fatalf("jsonbench: %v", err)
		}
	}

	// File read errors are always fatal, regardless of negation
	events, err := parseBenchFile(ts.MkAbs(filename))
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
	for i, event := range events {
		if i > 0 {
			fmt.Fprintln(&out)
		}
		writeEventGolden(&out, &event, sections)
	}

	fmt.Fprint(ts.Stdout(), trimLines(out.String()))
}

// CmdCmpBench compares two bench profile files by their deterministic fields.
// Usage: cmpbench <file1.json> <file2.json> [sections]
//
// Files MUST end with .json or .jsonl extension.
// Compares only Count and Gas values, ignoring timing data.
// The optional sections parameter is a comma-separated list (e.g., "opcodes,native").
// Fails if the deterministic fields differ.
// With negation (! cmpbench), expects files to differ.
func CmdCmpBench(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) < 2 || len(args) > 3 {
		ts.Fatalf("usage: cmpbench <file1.json> <file2.json> [sections]")
	}

	// Validate .json/.jsonl extension using centralized function
	for _, f := range args[:2] {
		if !benchops.IsJSONFormat(f) {
			ts.Fatalf("cmpbench: files must end with .json or .jsonl, got %q", f)
		}
	}

	var sections benchops.SectionFlags // default: 0 = all
	if len(args) == 3 {
		var err error
		sections, err = benchops.ParseSectionFlags(args[2])
		if err != nil {
			ts.Fatalf("cmpbench: %v", err)
		}
	}

	// File read errors are always fatal, regardless of negation
	formatted1, err := formatBenchFile(ts.MkAbs(args[0]), sections)
	if err != nil {
		ts.Fatalf("cmpbench: %s: %v", args[0], err)
	}

	formatted2, err := formatBenchFile(ts.MkAbs(args[1]), sections)
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
func formatBenchFile(path string, sections benchops.SectionFlags) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return FormatBenchOutput(data, sections)
}

// CmdBenchWithState implements the bench command with an external state.
// This allows callers to manage state separately from testscript env.Values.
func CmdBenchWithState(ts *testscript.TestScript, state *BenchState, args []string) {
	switch args[0] {
	case "start":
		if state.Running {
			ts.Fatalf("bench: profiler already running for this test (missing bench stop?)")
		}

		// Require exactly one argument: the filename
		if len(args) != 2 {
			ts.Fatalf("usage: bench start <filename.json>")
		}

		filename := args[1]

		// MUST be .json or .jsonl extension - use centralized validation
		if !benchops.IsJSONFormat(filename) {
			ts.Fatalf("bench start: filename must end with .json or .jsonl, got %q", filename)
		}

		// Acquire mutex to serialize access to global profiler across parallel tests
		benchMutex.Lock()
		state.mutexHeld = true

		state.OutputFile = filename
		state.Sections = 0 // Always all sections (no selection)

		benchops.Start()
		state.Running = true
		ts.Logf("bench: started profiling → %s", state.OutputFile)

	case "stop":
		if !state.Running {
			ts.Fatalf("bench: profiler not running (missing bench start?)")
		}

		results := benchops.Stop()
		state.Running = false

		// Release mutex only if we hold it
		if state.mutexHeld {
			benchMutex.Unlock()
			state.mutexHeld = false
		}

		// Write JSON to file in the work directory
		path := ts.MkAbs(state.OutputFile)
		if err := writeBenchResults(path, results); err != nil {
			ts.Fatalf("bench: %v", err)
		}

		state.Files = append(state.Files, state.OutputFile)
		ts.Logf("bench: stopped profiling, wrote %s", state.OutputFile)

		// Output golden format to stdout for automatic comparison
		var out strings.Builder
		results.WriteGolden(&out, 0) // 0 = all sections
		if _, err := fmt.Fprint(ts.Stdout(), trimLines(out.String())); err != nil {
			ts.Fatalf("bench: failed to write output: %v", err)
		}

	default:
		ts.Fatalf("bench: unknown subcommand %q (use 'start' or 'stop')", args[0])
	}
}

// parseBenchFile parses a bench profile file, auto-detecting the format.
func parseBenchFile(filename string) ([]benchops.BenchEvent, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseBenchData(data)
}

// writeEventGolden writes a benchops.BenchEvent in deterministic golden format.
// Uses benchops.Results.WriteGolden for the profile data.
func writeEventGolden(w *strings.Builder, event *benchops.BenchEvent, sections benchops.SectionFlags) {
	// Only print Package/Test if present (gno test format)
	if event.Package != "" {
		fmt.Fprintf(w, "Package: %s\n", event.Package)
	}
	if event.Test != "" {
		fmt.Fprintf(w, "Test: %s\n", event.Test)
	}

	// Use WriteGolden for deterministic profile output
	if event.Profile == nil {
		fmt.Fprintln(w, "<no profile>")
		return
	}

	event.Profile.WriteGolden(w, sections)
}

func trimLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}
