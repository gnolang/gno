package main

import (
	"flag"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type FlamegraphConfig struct {
	binPath   string
	duration  int
	output    string
}

func (c *FlamegraphConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.binPath, "bin", "", "path to the Go binary to profile (required)")
	fs.IntVar(&c.duration, "duration", 30, "duration of the profile in seconds")
	fs.StringVar(&c.output, "output", "flamegraph.svg", "output SVG file path")
}

func NewFlamegraphCmd(io commands.IO) *commands.Command {
	var cfg FlamegraphConfig

	return commands.NewCommand(
		commands.Metadata{
			Name: "flamegraph",
			ShortUsage: "gnodev flamegraph --bin <path> [--duration 30] [--output out.svg]",
			ShortHelp:  "Generate a flamegraph from a Go binary",
		},
		&cfg,
		func(_ context.Context, args []string) error {
			if cfg.binPath == "" {
				return errors.New("missing required flag: --bin")
			}
			return profile(cfg, io)
		},
	)
}

func profile(cfg FlamegraphConfig, io commands.IO) error {
	fmt.Fprintf(io.Out(), "Profiling binary: %s for %ds\n", cfg.binPath, cfg.duration)
	cmd := exec.Command(cfg.binPath)
	cmd.Env = append(os.Environ(), "PPROF_PROFILE=1")

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start binary: %w", err)
	}

	time.Sleep(time.Duration(cfg.duration) * time.Second)

	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	profilePath := "cpu.pprof"
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile file %s not found", profilePath)
	}
	return generateFlamegraph(profilePath, cfg.output, io)
}

func generateFlamegraph(profilePath, outputPath string, io commands.IO) error {
	fmt.Fprintln(io.Out(), "Generating flamegraph...")

	cmdA := exec.Command("go", "tool", "pprof", "-raw", profilePath)
	cmdB := exec.Command("/Users/moonia/FlameGraph/stackcollapse-go.pl") // tmp
	cmdC := exec.Command("/Users/moonia/FlameGraph/flamegraph.pl", "--width=2000", "--fontsize=10") // tmp

	aOut, err := cmdA.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pprof stdout pipe error: %w", err)
	}
	cmdB.Stdin = aOut

	bOut, err := cmdB.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stackcollapse-go.pl stdout pipe error: %w", err)
	}
	cmdC.Stdin = bOut

	var buf bytes.Buffer
	cmdC.Stdout = &buf

	if err := cmdC.Start(); err != nil {
		return fmt.Errorf("failed to start flamegraph.pl: %w", err)
	}
	if err := cmdB.Start(); err != nil {
		return fmt.Errorf("failed to start stackcollapse-go.pl: %w", err)
	}
	if err := cmdA.Run(); err != nil {
		return fmt.Errorf("pprof failed: %w", err)
	}
	if err := cmdB.Wait(); err != nil {
		return fmt.Errorf("stackcollapse-go.pl failed: %w", err)
	}
	if err := cmdC.Wait(); err != nil {
		return fmt.Errorf("flamegraph.pl failed: %w", err)
	}
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", outputPath, err)
	}

	fmt.Fprintf(io.Out(), "Flamegraph written to %s\n", outputPath)
	return nil
}
