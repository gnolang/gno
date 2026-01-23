//go:build gnobench

package integration

import (
	"fmt"
	"os"
	"path/filepath"

	gnointegration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"golang.org/x/tools/txtar"
)

// benchEnvKey is the key used to store BenchState in testscript.Env.Values.
type benchEnvKey struct{}

// benchSourceKey stores the source txtar file path for updating.
type benchSourceKey struct{}

// SetupGnolandBenchInMemory prepares gnoland testscript for benchops profiling
// in in-memory mode. This runs the node in the same process as the tests,
// allowing direct access to benchops profiling.
//
// When updateScripts is true, profile outputs are automatically written back
// to the txtar file as `-- <filename> --` sections.
//
// Note: This function assumes sequential test execution since benchops uses global state.
func SetupGnolandBenchInMemory(p *testscript.Params, testDir string, updateScripts bool) {
	// Build a list of txtar files
	txtarFiles, err := filepath.Glob(filepath.Join(testDir, "*.txtar"))
	if err != nil {
		// Pattern is malformed - this is a programmer error
		panic(fmt.Sprintf("invalid glob pattern for testDir %q: %v", testDir, err))
	}

	// Wrap setup to initialize bench state for each test
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		// Store bench state in env.Values for this test
		env.Values[benchEnvKey{}] = &gnointegration.BenchState{}

		// Find the source txtar file for this test by checking which txtar's
		// files exist in WorkDir. This works because Setup is called AFTER
		// testscript extracts the txtar files to WorkDir.
		sourcePath := findSourceTxtar(env.WorkDir, txtarFiles)
		env.Values[benchSourceKey{}] = sourcePath

		// Register deferred update of txtar with profile outputs
		if updateScripts && sourcePath != "" {
			env.Defer(func() {
				updateTxtarWithProfiles(env)
			})
		}

		return nil
	}

	// Register bench commands
	if p.Cmds == nil {
		p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	}

	p.Cmds["bench"] = func(ts *testscript.TestScript, neg bool, args []string) {
		if neg {
			ts.Fatalf("bench command does not support negation")
		}

		if len(args) == 0 {
			ts.Fatalf("usage: bench start [file] | bench stop")
		}

		state := getBenchState(ts)
		gnointegration.CmdBenchWithState(ts, state, args)
	}

	p.Cmds["jsonbench"] = gnointegration.CmdJSONBench
	p.Cmds["cmpbench"] = gnointegration.CmdCmpBench
}

// getBenchState retrieves the BenchState from the testscript environment.
func getBenchState(ts *testscript.TestScript) *gnointegration.BenchState {
	state, ok := ts.Value(benchEnvKey{}).(*gnointegration.BenchState)
	if !ok {
		ts.Fatalf("bench: state not found in env.Values (SetupGnolandBenchInMemory not called?)")
	}
	return state
}

// updateTxtarWithProfiles reads generated profile files, formats them,
// and updates the source txtar file with the formatted output.
func updateTxtarWithProfiles(env *testscript.Env) {
	state, ok := env.Values[benchEnvKey{}].(*gnointegration.BenchState)
	if !ok || len(state.Files) == 0 {
		return
	}

	sourcePath, ok := env.Values[benchSourceKey{}].(string)
	if !ok || sourcePath == "" {
		env.T().Log("bench: could not determine source txtar file")
		return
	}

	// Parse the source txtar
	archive, err := txtar.ParseFile(sourcePath)
	if err != nil {
		env.T().Log(fmt.Sprintf("bench: failed to parse txtar %s: %v", sourcePath, err))
		return
	}

	// Process each generated profile file
	for _, filename := range state.Files {
		profilePath := filepath.Join(env.WorkDir, filename)
		data, err := os.ReadFile(profilePath)
		if err != nil {
			env.T().Log(fmt.Sprintf("bench: failed to read profile %s: %v", profilePath, err))
			continue
		}

		// Format the profile output (deterministic fields only)
		formatted, err := gnointegration.FormatBenchOutput(data, state.Sections)
		if err != nil {
			env.T().Log(fmt.Sprintf("bench: failed to format profile %s: %v", filename, err))
			continue
		}

		// Update or add the file in the archive
		updateArchiveFile(archive, filename, []byte(formatted))
		env.T().Log(fmt.Sprintf("bench: updated %s in txtar", filename))
	}

	// Write the updated archive back
	if err := os.WriteFile(sourcePath, txtar.Format(archive), 0o644); err != nil {
		env.T().Log(fmt.Sprintf("bench: failed to write txtar %s: %v", sourcePath, err))
		return
	}

	env.T().Log(fmt.Sprintf("bench: updated txtar %s", sourcePath))
}

// findSourceTxtar tries to find which txtar file corresponds to the given WorkDir
// by checking which txtar's files exist in the WorkDir.
func findSourceTxtar(workDir string, txtarFiles []string) string {
	type match struct {
		path      string
		score     int
		fileCount int // number of files in the archive
	}

	var matches []match

	for _, src := range txtarFiles {
		archive, err := txtar.ParseFile(src)
		if err != nil {
			continue
		}

		// Count how many files from this archive exist in WorkDir
		score := 0
		for _, f := range archive.Files {
			checkPath := filepath.Join(workDir, f.Name)
			if _, err := os.Stat(checkPath); err == nil {
				score++
			}
		}

		matches = append(matches, match{
			path:      src,
			score:     score,
			fileCount: len(archive.Files),
		})
	}

	// First, try to find the txtar with the highest score (most matching files)
	var bestMatch string
	var bestScore int
	for _, m := range matches {
		if m.score > bestScore {
			bestScore = m.score
			bestMatch = m.path
		}
	}

	if bestMatch != "" {
		return bestMatch
	}

	// No files matched - this might be a txtar with no embedded files.
	// Find txtars with 0 files and use the one that hasn't been matched yet.
	var emptyTxtars []string
	for _, m := range matches {
		if m.fileCount == 0 {
			emptyTxtars = append(emptyTxtars, m.path)
		}
	}

	// If there's exactly one empty txtar, that's probably the match
	if len(emptyTxtars) == 1 {
		return emptyTxtars[0]
	}

	return ""
}

// updateArchiveFile updates or adds a file in the txtar archive.
func updateArchiveFile(archive *txtar.Archive, name string, data []byte) {
	// Ensure data ends with newline
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	// Look for existing file and update it
	for i := range archive.Files {
		if archive.Files[i].Name == name {
			archive.Files[i].Data = data
			return
		}
	}

	// Not found, append new file
	archive.Files = append(archive.Files, txtar.File{
		Name: name,
		Data: data,
	})
}
