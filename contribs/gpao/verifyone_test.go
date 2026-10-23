package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

// TestVerifyOneTakesABareMemPackage: verify-one's stdin is a contract with
// callers other than this oracle, which write the candidate's amino JSON and
// nothing after it. Given only that, the child fetches the imports itself.
func TestVerifyOneTakesABareMemPackage(t *testing.T) {
	const dep = "gno.land/p/nt/ufmt/v0"
	remote := nodeServing(t, dep)
	payload, err := amino.MarshalJSON(packageImporting(dep))
	require.NoError(t, err)

	self, err := os.Executable()
	require.NoError(t, err)
	cmd := exec.CommandContext(t.Context(), self,
		verifyOneCmdName, "-gno-root", gnoenv.RootDir(), "-remote", remote)
	cmd.Stdin = bytes.NewReader(payload)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	require.NoError(t, cmd.Run(), "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), childReadyMarker)
}
