package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// A valid BIP39 mnemonic (the standard gno integration test seed). The oracle
// only needs it to derive the approver address; no network access happens here.
const testMnemonic = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"

func newTestOracle(t *testing.T) *oracle {
	t.Helper()
	cfg := config{
		remote:    "http://127.0.0.1:26657", // not contacted in these tests
		chainID:   "test",
		mnemonic:  testMnemonic,
		gnoRoot:   gnoenv.RootDir(),
		gasFee:    defaultGasFee,
		gasWanted: defaultGasWanted,
	}
	o, err := newOracle(cfg, commands.NewTestIO())
	require.NoError(t, err)
	require.False(t, o.approver.IsZero(), "approver address must be derived")
	return o
}

// TestOracleTypecheckAcceptsValidPackage: a well-typed package importing a
// stdlib passes the off-chain typecheck — the oracle would approve it.
func TestOracleTypecheckAcceptsValidPackage(t *testing.T) {
	o := newTestOracle(t)

	const path = "gno.land/r/test/good"
	mpkg := &std.MemPackage{
		Name: "good",
		Path: path,
		Type: gno.MPUserProd,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
			{Name: "good.gno", Body: `package good

import "strings"

func Shout(cur realm, s string) string {
	return strings.ToUpper(s)
}`},
		},
	}

	require.NoError(t, o.typecheck(mpkg), "valid package must pass typecheck")
}

// TestOracleTypecheckRejectsInvalidPackage: an ill-typed package fails the
// off-chain typecheck — the oracle would NOT approve it (and the chain would
// reject it anyway).
func TestOracleTypecheckRejectsInvalidPackage(t *testing.T) {
	o := newTestOracle(t)

	const path = "gno.land/r/test/bad"
	mpkg := &std.MemPackage{
		Name: "bad",
		Path: path,
		Type: gno.MPUserProd,
		Files: []*std.MemFile{
			{Name: "bad.gno", Body: `package bad

func Boom(cur realm) string {
	return undefinedSymbol
}`},
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
		},
	}

	assert.Error(t, o.typecheck(mpkg), "ill-typed package must fail typecheck")
}
