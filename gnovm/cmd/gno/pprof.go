package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/profiler"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type pprofCfg struct{}

func newPprofCmd(io commands.IO) *commands.Command {
	cfg := &pprofCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "pprof",
			ShortUsage: "gno pprof [flags] <profile-file>",
			ShortHelp:  "interactive profiling analysis tool",
			LongHelp: `The pprof command provides an interactive interface for analyzing profile data,
similar to 'go tool pprof'. It supports various commands for viewing and filtering
profile information.

Examples:
  gno pprof cpu.prof           # Analyze CPU profile
  gno pprof -http=:8080 mem.prof  # Start web interface (future feature)

Interactive commands:
  top        - Show top functions by sample count
  list       - Show annotated source for a function
  tree       - Show call tree
  focus      - Focus on specific functions
  ignore     - Ignore functions in output
  hide       - Hide functions from output
  help       - Show help for commands
  quit       - Exit profiler

For detailed command help, use 'help <command>' in interactive mode.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execPprof(cfg, args, io)
		},
	)
}

func (c *pprofCfg) RegisterFlags(fs *flag.FlagSet) {
	// Future flags can be added here
	// fs.StringVar(&c.httpAddr, "http", "", "start HTTP server at this address")
}

func execPprof(cfg *pprofCfg, args []string, io commands.IO) error {
	if len(args) != 1 {
		return errors.New("usage: gno pprof <profile-file>")
	}

	profilePath := args[0]

	// Read profile data
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to read profile file: %w", err)
	}

	// Parse profile data (JSON format for now)
	var profile profiler.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("failed to parse profile data: %w", err)
	}

	// Create interactive CLI
	cli := profiler.NewProfilerCLI(&profile, nil)
	cli.SetInput(io.In())
	cli.SetOutput(io.Out())

	// Run interactive mode
	return cli.Run()
}

// Helper function to save profile data from Machine
func SaveProfile(m *gnolang.Machine, filename string) error {
	profile := m.StopProfiling()
	if profile == nil {
		return errors.New("no profiler data available")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Marshal profile to JSON
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}
