package mldsa65

import (
	"bytes"
	"crypto/subtle"
	"io"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
)

const (
	// PubKeySize is the size of a packed ML-DSA-65 public key (FIPS 204).
	PubKeySize = mldsa65.PublicKeySize // 1952
	// SignatureSize is the size of an ML-DSA-65 signature (FIPS 204).
	SignatureSize = mldsa65.SignatureSize // 3309
	// SeedSize is the size of the seed a key pair is expanded from.
	SeedSize = mldsa65.SeedSize // 32
)

// -------------------------------------

var _ crypto.PrivKey = PrivKeyMLDSA65{}

// PrivKeyMLDSA65 implements crypto.PrivKey. It holds the 32-byte FIPS 204
// seed ξ rather than the 4032-byte expanded key: the key pair is a pure
// function of the seed, and the seed is the form FIPS 204 recommends storing.
type PrivKeyMLDSA65 [SeedSize]byte

func (privKey PrivKeyMLDSA65) expand() (*mldsa65.PublicKey, *mldsa65.PrivateKey) {
	seed := [SeedSize]byte(privKey)
	return mldsa65.NewKeyFromSeed(&seed)
}

// Bytes marshals the privkey using amino encoding w/ type information.
func (privKey PrivKeyMLDSA65) Bytes() []byte {
	return amino.MustMarshalAny(privKey)
}

// Sign produces a deterministic ML-DSA-65 signature on msg, with an empty
// context string. Deterministic signing matches ed25519 and keeps signatures
// reproducible; the hedged variant can be adopted later without any change on
// the verifying side.
func (privKey PrivKeyMLDSA65) Sign(msg []byte) ([]byte, error) {
	_, sk := privKey.expand()
	sig := make([]byte, SignatureSize)
	if err := mldsa65.SignTo(sk, msg, nil, false, sig); err != nil {
		return nil, err
	}
	return sig, nil
}

// PubKey derives the public key from the seed.
func (privKey PrivKeyMLDSA65) PubKey() crypto.PubKey {
	pk, _ := privKey.expand()
	var buf [PubKeySize]byte
	pk.Pack(&buf)
	return PubKeyMLDSA65(buf)
}

// Equals runs in constant time.
func (privKey PrivKeyMLDSA65) Equals(other crypto.PrivKey) bool {
	otherKey, ok := other.(PrivKeyMLDSA65)
	return ok && subtle.ConstantTimeCompare(privKey[:], otherKey[:]) == 1
}

// GenPrivKey generates a new ML-DSA-65 private key from crypto.CReader().
func GenPrivKey() PrivKeyMLDSA65 {
	return genPrivKey(crypto.CReader())
}

func genPrivKey(rand io.Reader) PrivKeyMLDSA65 {
	var privKey PrivKeyMLDSA65
	if _, err := io.ReadFull(rand, privKey[:]); err != nil {
		panic(err)
	}
	return privKey
}

// GenPrivKeyFromSecret hashes the secret with SHA2 and uses the 32-byte
// output as the seed, like ed25519.GenPrivKeyFromSecret.
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyFromSecret(secret []byte) PrivKeyMLDSA65 {
	return PrivKeyMLDSA65(crypto.Sha256(secret))
}

// -------------------------------------

var _ crypto.PubKey = PubKeyMLDSA65{}

// PubKeyMLDSA65 implements crypto.PubKey for ML-DSA-65.
type PubKeyMLDSA65 [PubKeySize]byte

// Address is the SHA256-20 of the raw pubkey bytes, as for the other keys.
func (pubKey PubKeyMLDSA65) Address() crypto.Address {
	return crypto.MustAddressFromBytes(tmhash.SumTruncated(pubKey[:]))
}

// Bytes marshals the PubKey using amino encoding.
func (pubKey PubKeyMLDSA65) Bytes() []byte {
	return amino.MustMarshalAny(pubKey)
}

// VerifyBytes reports whether sig is a valid ML-DSA-65 signature of msg,
// with an empty context string, under pubKey.
func (pubKey PubKeyMLDSA65) VerifyBytes(msg []byte, sig []byte) bool {
	if len(sig) != SignatureSize {
		return false
	}
	var pk mldsa65.PublicKey
	buf := [PubKeySize]byte(pubKey)
	pk.Unpack(&buf)
	return mldsa65.Verify(&pk, msg, nil, sig)
}

func (pubKey PubKeyMLDSA65) String() string {
	return crypto.PubKeyToBech32(pubKey)
}

func (pubKey PubKeyMLDSA65) Equals(other crypto.PubKey) bool {
	otherKey, ok := other.(PubKeyMLDSA65)
	return ok && bytes.Equal(pubKey[:], otherKey[:])
}
