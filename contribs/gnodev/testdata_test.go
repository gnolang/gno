package main

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/stretchr/testify/require"
)

func TestScripts(t *testing.T) {
	p := integration.NewTestingParams(t, "testdata")

	buildDir := t.TempDir()
	gnodevBin := filepath.Join(buildDir, "gnodev")
	buildCmd := exec.Command(
		"go",
		"build",
		"-ldflags",
		"-X github.com/gnolang/gno/tm2/pkg/version.Version=testscript-version",
		"-o",
		gnodevBin,
		".",
	)
	buildCmd.Dir = "."
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, string(output))

	integration.RegisterExecCommand(&p, "gnodev", gnodevBin)
	integration.RunTestscript(t, p)
}
