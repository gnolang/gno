package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStagingHomeMarkdown(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		chainID     string
		rpc         string
		contains    []string
		notContains []string
	}{
		{
			name:    "names the chain and the real network",
			chainID: "acme-preview",
			rpc:     "https://rpc.acme.test",
			contains: []string{
				"**Chain ID**: `acme-preview`",
				"**RPC**: `https://rpc.acme.test`",
				"not a gno.land network",
				mainnetURL,
				mainnetChainID,
				// The divergences that decide whether a deployment works.
				"namespace enforcement is off",
				"code_submission_policy=inert",
				// Keys do carry across chains, so this is advice, not a promise.
				"Use a test key here",
			},
		},
		{
			name:    "no operator RPC leaves the row and the flag out",
			chainID: "dev",
			rpc:     "",
			// Printing appcfg.RemoteHelp here would emit the node's own listen
			// address and put 127.0.0.1 in a command a visitor is meant to copy.
			notContains: []string{"**RPC**:", "-remote "},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := stagingHomeMarkdown(tc.chainID, tc.rpc)
			for _, want := range tc.contains {
				assert.Contains(t, got, want)
			}
			for _, unwanted := range tc.notContains {
				assert.NotContains(t, got, unwanted)
			}
		})
	}
}

// TestStagingHomeMarkdownDeployCommandTargetsThisChain guards what makes the
// page useful rather than decorative: the command a visitor copies has to point
// at the chain they are looking at, with gas that actually covers an addpkg.
func TestStagingHomeMarkdownDeployCommandTargetsThisChain(t *testing.T) {
	t.Parallel()

	got := stagingHomeMarkdown("weird-chain-9", "https://rpc.weird.test")

	assert.Contains(t, got, "-broadcast -chainid weird-chain-9 -remote https://rpc.weird.test")
	// A five-line realm measured 4_282_275 gas on a live gnodev chain, so the
	// 2_000_000 from the maketx call examples would fail before it deployed.
	assert.Contains(t, got, "-gas-wanted 20000000")
	assert.Greater(t, addpkgGasWanted, 4_282_275)
}

// TestStagingHomeMarkdownChainIDIsNotIncidentallyPresent: asserting on a bare
// chain-id like "dev" passes no matter what the function does, because the page
// says "gnodev". Pin the rendered form instead.
func TestStagingHomeMarkdownChainIDIsNotIncidentallyPresent(t *testing.T) {
	t.Parallel()

	got := stagingHomeMarkdown("dev", "")

	assert.Contains(t, got, "**Chain ID**: `dev`")
	assert.True(t, strings.Contains(got, "gnodev"), "the page does mention gnodev, which is why a bare \"dev\" assertion is vacuous")
}
