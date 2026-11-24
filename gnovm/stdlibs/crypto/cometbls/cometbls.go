package cometbls

import (
	"reflect"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// X_verifyZKP is the host-side native binding for the `verifyZKP` function
// declared in cometbls.gno. It returns an empty string on successful proof
// verification and a non-empty error message otherwise — gno cannot currently
// return a Go `error` across the native ABI, so the gno wrapper converts the
// string back into an error value.
//
// `headerEncoded` must be exactly EncodedLightHeaderSize bytes (116); see
// EncodeLightHeader in cometbls.gno for the canonical encoder.
func X_verifyZKP(m *gno.Machine, chainID string, trustedValidatorsHash gno.TypedValue, headerEncoded []byte, zkp gno.TypedValue) string {
	// The row is flat, so each operand is copied up to one byte past the one
	// length it can have, which keeps the error verifyZKP returns.
	return verifyZKP(chainID, prefix(m, trustedValidatorsHash, 32+1), headerEncoded, prefix(m, zkp, ExpectedProofSize+1))
}

// verifyZKP returns VerifyZKP's error message, or "" when the proof verifies.
func verifyZKP(chainID string, trustedValidatorsHash []byte, headerEncoded []byte, zkp []byte) string {
	if err := VerifyZKP(chainID, trustedValidatorsHash, headerEncoded, zkp); err != nil {
		return err.Error()
	}
	return ""
}

// prefix returns the first n bytes of the Gno []byte tv, or all of it when
// shorter.
func prefix(m *gno.Machine, tv gno.TypedValue, n int) []byte {
	tv.DeepFill(m.Store)
	if tv.GetLength() > n {
		tv = tv.GetSlice(nil, 0, n)
	}
	var b []byte
	gno.Gno2GoValue(&tv, reflect.ValueOf(&b).Elem())
	return b
}
