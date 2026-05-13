package crypto_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// TestCodecParity_Crypto asserts the cross-codec parity invariant for the
// genproto2-registered types across every tm2/pkg/crypto/* subpackage.
//
// Exercises surfaces the other parity arrays miss:
//   - typed byte arrays (PubKeyEd25519 = [32]byte, etc.)
//   - nested byte slices ([][]byte in SimpleProof.Aunts)
//   - interface slices with Any-wrapped elements
//     ([]crypto.PubKey in PubKeyMultisigThreshold.PubKeys, which forces
//     every element through MarshalAnyBinary2 and the genproto2 Any-
//     wrapping generator path)
//
// Add new cases by appending to parityCasesCrypto.
func TestCodecParity_Crypto(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	// Register every package whose types appear in the fixture array, plus
	// any package whose concrete types need to be resolvable through the
	// crypto.PubKey interface (so multisig's []PubKey can Any-wrap them).
	cdc.RegisterPackage(ed25519.Package)
	cdc.RegisterPackage(secp256k1.Package)
	cdc.RegisterPackage(multisig.Package)
	cdc.RegisterPackage(hd.Package)
	cdc.RegisterPackage(merkle.Package)
	cdc.Seal()

	for i, c := range parityCasesCrypto() {
		c := c
		name := fmt.Sprintf("%d/%s", i, c.name)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}

func parityCasesCrypto() []struct {
	name string
	v    any
} {
	// Deterministic byte patterns — no PRNG so the fixture is reproducible.
	pkEd := ed25519.PubKeyEd25519{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}
	pvEd := ed25519.PrivKeyEd25519{}
	for i := range pvEd {
		pvEd[i] = byte(i)
	}

	pkSec := secp256k1.PubKeySecp256k1{}
	for i := range pkSec {
		pkSec[i] = byte(i + 100)
	}
	pvSec := secp256k1.PrivKeySecp256k1{}
	for i := range pvSec {
		pvSec[i] = byte(i + 200)
	}

	// Multisig 2-of-3 with heterogeneous PubKey concrete types. Exercises
	// the interface-slice / Any-wrapping encoder path.
	pkEd2 := ed25519.PubKeyEd25519{0x99, 0xaa, 0xbb, 0xcc}
	multi := &multisig.PubKeyMultisigThreshold{
		K: 2,
		PubKeys: []crypto.PubKey{
			pkEd,
			pkSec,
			pkEd2,
		},
	}
	// Same shape with a nil PubKey entry — exercises Any-wrapping over
	// a nil interface in the middle of a slice. The encoder writes 0x00
	// for the nil element; the decoder must recover it as a nil
	// crypto.PubKey interface.
	multiWithNil := &multisig.PubKeyMultisigThreshold{
		K: 2,
		PubKeys: []crypto.PubKey{
			pkEd,
			nil,
			pkSec,
		},
	}

	// BIP44Params with non-trivial content.
	bip := &hd.BIP44Params{
		Purpose:      44,
		CoinType:     118,
		Account:      0,
		Change:       false,
		AddressIndex: 7,
	}

	// SimpleProof with non-trivial Aunts (the [][]byte case).
	simpleProof := &merkle.SimpleProof{
		Total: 4,
		Index: 1,
		LeafHash: []byte{
			0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11,
		},
		Aunts: [][]byte{
			{0x01, 0x02, 0x03},
			{0x04, 0x05, 0x06, 0x07},
			{0x08},
		},
	}

	// Proof composed of multiple ProofOp entries.
	proof := &merkle.Proof{
		Ops: []merkle.ProofOp{
			{Type: "iavl:v", Key: []byte("key1"), Data: []byte{0xde, 0xad}},
			{Type: "multistore", Key: []byte("key2"), Data: []byte{0xbe, 0xef}},
		},
	}

	// Single ProofOp.
	proofOp := &merkle.ProofOp{
		Type: "iavl:v",
		Key:  []byte("hello"),
		Data: []byte{0xfe, 0xed, 0xfa, 0xce},
	}

	// SimpleProofNode is recursive (*SimpleProofNode fields) — build a
	// depth-2 tree with one sibling on each side.
	leaf := &merkle.SimpleProofNode{Hash: []byte{0xaa}}
	// Note: cyclic Parent pointers can't roundtrip via amino's
	// tree-oriented encoding; keep the tree DAG-like.
	spn := &merkle.SimpleProofNode{
		Hash: []byte{0xff, 0xee, 0xdd},
		Left: leaf,
	}

	return []struct {
		name string
		v    any
	}{
		{"PubKeyEd25519", &pkEd},
		{"PrivKeyEd25519", &pvEd},
		{"PubKeySecp256k1", &pkSec},
		{"PrivKeySecp256k1", &pvSec},
		{"BIP44Params", bip},
		{"SimpleProof", simpleProof},
		{"ProofOp", proofOp},
		{"Proof", proof},
		{"SimpleProofNode", spn},
		{"PubKeyMultisigThreshold", multi},
		{"PubKeyMultisigThreshold/with-nil", multiWithNil},
	}
}
