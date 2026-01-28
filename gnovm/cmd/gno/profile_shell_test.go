package main

import (
	"bytes"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/profiler"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func TestExecuteProfileCommand(t *testing.T) {
	// Create a minimal profile for testing
	prof := profiler.NewProfiler(profiler.ProfileCPU, 1)
	prof.StartProfiling(nil, profiler.Options{Type: profiler.ProfileCPU, SampleRate: 1})

	// Stop immediately to get an empty but valid profile
	profile := prof.StopProfiling()

	cfg := &test.ProfileConfig{}
	store := test.NewStoreAdapter(gno.NewStore(nil, nil, nil))

	tests := []struct {
		name     string
		command  string
		checkOut func(t *testing.T, out, err string)
	}{
		{
			name:    "help command",
			command: "help",
			checkOut: func(t *testing.T, out, err string) {
				t.Helper()
				require.Contains(t, err, "Commands:")
				require.Contains(t, err, "help")
				require.Contains(t, err, "text")
				require.Contains(t, err, "top")
				require.Contains(t, err, "list")
				require.Contains(t, err, "clear")
				require.Contains(t, err, "exit")
			},
		},
		{
			name:    "exit command",
			command: "exit",
			checkOut: func(t *testing.T, out, err string) {
				t.Helper()
				require.Contains(t, err, "Exiting profiler shell")
			},
		},
		{
			name:    "unknown command",
			command: "foobar",
			checkOut: func(t *testing.T, out, err string) {
				t.Helper()
				require.Contains(t, err, "unknown command")
			},
		},
		{
			name:    "text command",
			command: "text",
			checkOut: func(t *testing.T, out, err string) {
				t.Helper()
				require.Contains(t, out, "Profile Type:")
			},
		},
		{
			name:    "json command",
			command: "json",
			checkOut: func(t *testing.T, out, err string) {
				t.Helper()
				require.Contains(t, out, "{")
				require.Contains(t, out, "}")
			},
		},
		{
			name:    "clear command",
			command: "clear",
			checkOut: func(t *testing.T, out, err string) {
				t.Helper()
				// Check for ANSI clear screen sequence in stderr
				require.Contains(t, err, "\033[H\033[2J")
				require.Contains(t, err, "Profiler shell ready")
				require.Empty(t, out)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := bytes.NewBuffer(nil)
			errBuf := bytes.NewBuffer(nil)

			io := commands.NewTestIO()
			io.SetIn(bytes.NewBufferString(""))
			io.SetOut(commands.WriteNopCloser(outBuf))
			io.SetErr(commands.WriteNopCloser(errBuf))

			executeProfileCommand(tt.command, io, profile, cfg, store)

			tt.checkOut(t, outBuf.String(), errBuf.String())
		})
	}
}
