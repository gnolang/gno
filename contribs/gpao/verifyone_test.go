package main

import (
	"bytes"
	"os"
	"os/exec"
	"slices"
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

// TestVerifyOneUnreadableInputIsNotAVerdict: the parent writes verify-one's
// stdin, so input the child cannot read is this oracle's fault, never the
// candidate's. Exit 1 is a rejection, so it must not be what reports it.
func TestVerifyOneUnreadableInputIsNotAVerdict(t *testing.T) {
	candidate, err := amino.MarshalJSON(packageImporting("gno.land/p/nt/ufmt/v0"))
	require.NoError(t, err)
	cases := map[string][]byte{
		"unreadable candidate":  []byte("not json"),
		"undecodable candidate": []byte(`{"files":5}`),
		"unreadable seed":       append(slices.Clone(candidate), []byte(" not json")...),
		"undecodable seed":      append(slices.Clone(candidate), []byte(` {"imports":5}`)...),
	}
	self, err := os.Executable()
	require.NoError(t, err)
	for name, stdin := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), self,
				verifyOneCmdName, "-gno-root", gnoenv.RootDir(), "-remote", "http://127.0.0.1:1")
			cmd.Stdin = bytes.NewReader(stdin)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			var exit *exec.ExitError
			require.ErrorAs(t, cmd.Run(), &exit)
			assert.Equal(t, exitResolverUnavailable, exit.ExitCode(), "stderr: %s", stderr.String())
		})
	}
}
