package main

import (
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func TestSetupWeb(t *testing.T) {
	opts := defaultWebOptions
	opts.bind = "127.0.0.1:0" // random port
	stdio := commands.NewDefaultIO()

	// Open /dev/null as a write-only file
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	require.NoError(t, err)
	defer devNull.Close()

	stdio.SetOut(devNull)

	_, err = setupWeb(&opts, []string{}, stdio)
	require.NoError(t, err)
}
