//go:build !libsecp256k1

package secp256k1

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// Sign creates an ECDSA signature on curve Secp256k1, using SHA256 on the msg.
// The returned signature will be of the form R || S (in lower-S form).
func (privKey PrivKeySecp256k1) Sign(msg []byte) ([]byte, error) {
	priv, _ := btcec.PrivKeyFromBytes(privKey[:])

	sig, err := ecdsa.SignCompact(priv, crypto.Sha256(msg), false) // ref uncompressed pubkey
	if err != nil {
		return nil, err
	}

	// remove compact sig recovery code byte at the beginning
	return sig[1:], nil
}

// VerifyBytes verifies a signature of the form R || S.
// It rejects signatures which are not in lower-S form.
func (pubKey PubKeySecp256k1) VerifyBytes(msg []byte, sigStr []byte) bool {
	if len(sigStr) != 64 {
		return false
	}

	pub, err := secp256k1.ParsePubKey(pubKey[:])
	if err != nil {
		return false
	}

	psig, ok := signatureFromBytes(sigStr)
	if !ok {
		return false
	}

	return psig.Verify(crypto.Sha256(msg), pub)
}

// Read Signature struct from R || S. Caller needs to ensure
// that len(sigStr) == 64.
func signatureFromBytes(sigStr []byte) (*ecdsa.Signature, bool) {
	// parse the signature:
	var r, s secp256k1.ModNScalar
	if r.SetByteSlice(sigStr[:32]) {
		return nil, false // overflow
	}
	if s.SetByteSlice(sigStr[32:]) {
		return nil, false
	}

	// Reject malleable signatures. libsecp256k1 does this check but btcec doesn't.
	if s.IsOverHalfOrder() {
		return nil, false
	}

	return ecdsa.NewSignature(&r, &s), true
}
