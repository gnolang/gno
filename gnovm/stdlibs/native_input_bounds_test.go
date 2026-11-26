package stdlibs_test

import (
	"fmt"
	"io"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/stretchr/testify/require"
)

// Each oversized operand is bounded before its Gno-to-Go copy, which no gas
// pays for: in the native for ed25519 and cometbls, in the .gno wrapper for
// bn254, modexp and secp256k1.RecoverPubkey. See adr/native_input_bounds.md.
func TestNativeWrapperBoundsCopy(t *testing.T) {
	opts := newOpts(t)

	// allocated returns the host bytes allocated by a main that runs setup
	// once, then calls call four times, on slices of one 16 MiB buffer.
	allocated := func(t *testing.T, imp, setup, call string) uint64 {
		t.Helper()
		src := fmt.Sprintf(`package main

import %q

func main() {
	buf := make([]byte, 1<<24)
	%s
	for i := 0; i < 4; i++ {
		_ = %s
	}
}
`, imp, setup, call)
		return filetestAlloc(t, opts, "bounds.gno", src, call)
	}

	const header = "cometbls.LightHeader{ValidatorsHash: buf[:32], NextValidatorsHash: buf[:32], AppHash: buf[:32]}"
	// chainID is left unbounded: the dispatcher hands the native the Gno
	// string itself, so a long one built once is not copied per call.
	const chainID = "s := string(buf[:%s])"
	for _, tc := range []struct {
		imp, operand, setup, call, honest string
	}{
		{"crypto/ed25519", "publicKey", "", "ed25519.Verify(buf[:%s], nil, buf[:64])", "32"},
		{"crypto/ed25519", "signature", "", "ed25519.Verify(buf[:32], nil, buf[:%s])", "64"},
		{"crypto/cometbls", "trustedValidatorsHash", "", `cometbls.VerifyZKP("x", buf[:%s], ` + header + `, buf[:384])`, "32"},
		{"crypto/cometbls", "zkp", "", `cometbls.VerifyZKP("x", buf[:32], ` + header + `, buf[:%s])`, "384"},
		{"crypto/cometbls", "chainID", chainID, `cometbls.VerifyZKP(s, buf[:32], ` + header + `, buf[:384])`, "31"},
		{"crypto/bn254", "G1Add", "", "bn254.G1Add(buf[:%s])", "128"},
		{"crypto/modexp", "base", "", "modexp.ModExp(buf[:%s], buf[:1], buf[:1])", "1024"},
		// RecoverPubkey returns two values; the literal hands the loop one.
		{"crypto/secp256k1", "digest", "", "func() bool { _, ok := secp256k1.RecoverPubkey(buf[:%s], buf[:65]); return ok }()", "32"},
		{"crypto/secp256k1", "signature", "", "func() bool { _, ok := secp256k1.RecoverPubkey(buf[:32], buf[:%s]); return ok }()", "65"},
	} {
		t.Run(path.Base(tc.imp)+"/"+tc.operand, func(t *testing.T) {
			fill := func(s, n string) string { return strings.ReplaceAll(s, "%s", n) }
			honest := fill(tc.call, tc.honest)
			oversized := fill(tc.call, "1<<24")
			honestSetup := fill(tc.setup, tc.honest)
			oversizedSetup := fill(tc.setup, "1<<24")
			allocated(t, tc.imp, honestSetup, honest) // loads the package into the store
			honestAlloc := allocated(t, tc.imp, honestSetup, honest)
			oversizedAlloc := allocated(t, tc.imp, oversizedSetup, oversized)
			// Unbounded, the four calls copy 64 MiB more than the honest ones.
			limit := honestAlloc + 1<<24
			if tc.setup != "" {
				limit += 1 << 24 // the setup copies the buffer once
			}
			require.Less(t, oversizedAlloc, limit,
				"%s allocated %d bytes against %d for %s",
				strings.TrimSpace(oversizedSetup+" "+oversized), oversizedAlloc, honestAlloc,
				strings.TrimSpace(honestSetup+" "+honest))
		})
	}
}

// A slice built by a variadic call is List-backed, so its Gno-to-Go copy goes
// element by element rather than as one block. The bounds hold the same.
func TestNativeWrapperBoundsCopyListBacked(t *testing.T) {
	opts := newOpts(t)

	const n = 100000
	elems := strings.Repeat("0, 0, 0, 0, 0, 0, 0, 0, 0, 0,\n", n/10)
	// allocated returns the host bytes allocated by a main that builds an
	// n-element List-backed slice, then calls call loops times.
	allocated := func(t *testing.T, imp, call string, loops int) uint64 {
		t.Helper()
		src := fmt.Sprintf(`package main

import %q

func mk(b ...byte) []byte { return b }

func main() {
	list := mk(%s)
	h := make([]byte, 32)
	p := make([]byte, 384)
	_, _, _ = list, h, p
	for i := 0; i < %d; i++ {
		_ = %s
	}
}
`, imp, elems, loops, call)
		return filetestAlloc(t, opts, "bounds_list.gno", src, call)
	}

	const loops = 200
	const header = "cometbls.LightHeader{ValidatorsHash: h, NextValidatorsHash: h, AppHash: h}"
	for _, tc := range []struct{ imp, operand, call string }{
		{"crypto/ed25519", "publicKey", "ed25519.Verify(list, nil, p[:64])"},
		{"crypto/ed25519", "signature", "ed25519.Verify(h, nil, list)"},
		{"crypto/cometbls", "trustedValidatorsHash", `cometbls.VerifyZKP("x", list, ` + header + `, p)`},
		{"crypto/cometbls", "zkp", `cometbls.VerifyZKP("x", h, ` + header + `, list)`},
	} {
		t.Run(path.Base(tc.imp)+"/"+tc.operand, func(t *testing.T) {
			allocated(t, tc.imp, tc.call, 1) // loads the package into the store
			none := allocated(t, tc.imp, tc.call, 0)
			calls := allocated(t, tc.imp, tc.call, loops)
			perCall := (int64(calls) - int64(none)) / loops
			// Unbounded, each call copies the n elements, about n bytes or more.
			require.Less(t, perCall, int64(n/4),
				"%s allocated %d host bytes per call on a %d-element List-backed operand", tc.call, perCall, n)
		})
	}
}

// filetestAlloc returns the host bytes allocated while opts runs src as the
// filetest name; msg labels a failed run.
func filetestAlloc(t *testing.T, opts *test.TestOptions, name, src, msg string) uint64 {
	t.Helper()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, _, _, err := opts.RunFiletest(name, []byte(src), opts.TestStore) //nolint:dogsled // only the error matters here.
	runtime.ReadMemStats(&after)
	require.NoError(t, err, msg)
	return after.TotalAlloc - before.TotalAlloc
}

// newOpts returns test options over this repository's stdlibs.
func newOpts(t *testing.T) *test.TestOptions {
	t.Helper()
	rootDir, err := filepath.Abs("../..")
	require.NoError(t, err)
	return test.NewTestOptions(rootDir, io.Discard, io.Discard, nil)
}

// runMain runs src as a filetest and fails t if its main panics.
func runMain(t *testing.T, src string) {
	t.Helper()
	filetestAlloc(t, newOpts(t), "result.gno", src, "main panicked")
}

// A public key of the wrong length fails verification rather than panicking.
func TestVerifyWrongKeyLength(t *testing.T) {
	runMain(t, `package main

import (
	"crypto/ed25519"
	"strconv"
)

func main() {
	signature := make([]byte, 64)
	for _, n := range []int{0, 31, 33} {
		if ed25519.Verify(make([]byte, n), nil, signature) {
			panic("verified a " + strconv.Itoa(n) + "-byte public key")
		}
	}
}
`)
}

// A signature of the wrong length fails, even one whose first 64 bytes verify.
func TestVerifyWrongSignatureLength(t *testing.T) {
	runMain(t, `package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"strconv"
)

func main() {
	publicKey, _ := hex.DecodeString("0D853FA898A07EB91F618BB3E8B738B0E45BE9B3053799A2C42F8204F5FA3505")
	signature, _ := hex.DecodeString("2B39638983858715AD2FA059665ADFE267936B8F20C4DA01E9650958E0CA65C0255C75164360F468087FE8385140E48EE3471E332472A50AEE599F9D0EADD106")
	for _, sig := range [][]byte{nil, signature[:63], append(signature, 0)} {
		if ed25519.Verify(publicKey, []byte("hello gno.land"), sig) {
			panic("verified a " + strconv.Itoa(len(sig)) + "-byte signature")
		}
	}
}
`)
}

// An oversized operand fails with the error the native gives it, in the
// native's order: a negative height is reported before a long proof.
func TestVerifyZKPOversizedOperands(t *testing.T) {
	runMain(t, `package main

import "crypto/cometbls"

func main() {
	hash := make([]byte, 32)
	long := make([]byte, 1024)
	header := cometbls.LightHeader{
		ValidatorsHash: hash, NextValidatorsHash: hash, AppHash: hash,
	}
	for _, tc := range []struct {
		height   int64
		tvh, zkp []byte
		want     string
	}{
		{0, long, long, "cometbls: invalid commitment"},
		{-1, hash, long, "cometbls: invalid height"},
		{0, hash, long, "cometbls: invalid raw proof"},
	} {
		header.Height = tc.height
		err := cometbls.VerifyZKP("x", tc.tvh, header, tc.zkp)
		if err == nil || err.Error() != tc.want {
			panic("want " + tc.want)
		}
	}
}
`)
}
