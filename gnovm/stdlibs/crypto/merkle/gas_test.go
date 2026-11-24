package merkle_test

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/stretchr/testify/require"
)

// runGas returns the gas of a main passing slices of one 16 MiB buffer to
// call, so two runs differ only by what call charged for the slice lengths.
func runGas(t *testing.T, opts *test.TestOptions, call string) int64 {
	t.Helper()
	src := fmt.Sprintf(`package main

import "crypto/merkle"

func main() {
	buf := make([]byte, 1<<24)
	_ = %s
}
`, call)
	_, gas, _, err := opts.RunFiletest("merkle_gas.gno", []byte(src), opts.TestStore)
	require.NoError(t, err)
	return gas
}

func newOpts(t *testing.T) *test.TestOptions {
	t.Helper()
	rootDir, err := filepath.Abs("../../../..")
	require.NoError(t, err)
	opts := test.NewTestOptions(rootDir, io.Discard, io.Discard, nil)
	// Warm-up: the first run pays to load crypto/merkle into the store.
	runGas(t, opts, "merkle.LeafHash(buf[:0])")
	return opts
}

// Each byte VerifySimpleProof and InnerHash hash costs LeafHash's rate or more.
func TestMerkleHashedBytesCharged(t *testing.T) {
	opts := newOpts(t)
	// charged is the gas of a 1 MiB slice over an empty one; "1<<20" and
	// "0<<20" keep both sources the same length.
	charged := func(call string) int64 {
		return runGas(t, opts, fmt.Sprintf(call, "1<<20")) -
			runGas(t, opts, fmt.Sprintf(call, "0<<20"))
	}
	leafRate := charged("merkle.LeafHash(buf[:%s])")
	require.Positive(t, leafRate)

	for _, tc := range []struct {
		call   string
		hashes int64
	}{
		// verifySimpleProof hashes the leaf twice: for the proof, then in Verify.
		{"merkle.VerifySimpleProof(buf[:32], buf[:%s], 0, 1, nil)", 2},
		{"merkle.InnerHash(buf[:%s], nil)", 1},
		{"merkle.InnerHash(nil, buf[:%s])", 1},
	} {
		got := charged(tc.call)
		require.GreaterOrEqual(t, got, tc.hashes*leafRate,
			"%s: 1 MiB charged %d gas, LeafHash charges %d per hash of it", tc.call, got, leafRate)
	}
}

// A root longer than 32 bytes is rejected before the native copies it, so a
// 16 MiB root allocates about what a 32-byte one does.
func TestVerifySimpleProofOversizedRootNotCopied(t *testing.T) {
	opts := newOpts(t)
	honest := "merkle.VerifySimpleProof(buf[:32], buf[:12], 0, 1, nil)"
	oversized := "merkle.VerifySimpleProof(buf[:1<<24], buf[:12], 0, 1, nil)"

	allocated := func(call string) uint64 {
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		runGas(t, opts, call)
		runtime.ReadMemStats(&after)
		return after.TotalAlloc - before.TotalAlloc
	}
	honestAlloc := allocated(honest)
	oversizedAlloc := allocated(oversized)

	require.Less(t, oversizedAlloc, honestAlloc+(1<<23),
		"a 16 MiB rootHash allocated %d bytes against %d for a 32-byte one", oversizedAlloc, honestAlloc)
}

// A real root verifies at 32 bytes and at no other length; runGas fails on
// the panic, whose value is the length of a root that verified.
func TestVerifySimpleProofRootLength(t *testing.T) {
	runGas(t, newOpts(t), `func() bool {
		root := merkle.LeafHash(buf[:1])
		if !merkle.VerifySimpleProof(root, buf[:1], 0, 1, nil) {
			panic("the 32-byte root was rejected")
		}
		for _, bad := range [][]byte{root[:0], root[:31], append(root, 0), append(root, buf[:32]...)} {
			if merkle.VerifySimpleProof(bad, buf[:1], 0, 1, nil) {
				panic(len(bad))
			}
		}
		return false
	}()`)
}
