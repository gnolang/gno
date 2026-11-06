// Package evmnet verifies drand "evmnet" beacons off-chain.
//
// evmnet is the drand network that signs on BN254 (scheme
// bls-bn254-unchained-on-g1), which is the curve gno.land exposes as native
// precompiles. The relayer uses this package to check a beacon before paying
// gas to submit it. It is also the reference the gno port in
// gno.land/p/drand/v0 is tested against: its vectors_test.gno was
// produced with HashToG1 and Verify from this package.
package evmnet

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/drand/kyber"
	"github.com/drand/kyber/pairing/bn254"
	"github.com/drand/kyber/sign/bls" //nolint:staticcheck // deprecated for aggregation only; we verify single threshold signatures, as drand does
	"golang.org/x/crypto/sha3"
)

// Network parameters, from https://api.drand.sh/<ChainHash>/info.
const (
	ChainHash    = "04f1e9062b8a81f848fded9c12306733282b2727ecced50032187751166ec8c3"
	PublicKeyHex = "07e1d1d335df83fa98462005690372c643340060d205306a9aa8106b6bd0b3820557ec32c2ad488e4d4f6008f89a346f18492092ccc0d594610de2732c8b808f0095685ae3a85ba243747b1b2f426049010f6b73a0cf1d389351d5aaaa1047f6297d3a4f9749b33eb2d904c9d9ebf17224150ddd7abd7567a9bec6c74480ee0b"
	GenesisTime  = int64(1727521075)
	Period       = int64(3)
	DST          = "BLS_SIG_BN254G1_XMD:KECCAK-256_SVDW_RO_NUL_"
	SigSize      = 64
)

var (
	suite  = newSuite()
	scheme = bls.NewSchemeOnG1(suite)
	pubKey = mustPubKey()
)

func newSuite() *bn254.Suite {
	s := bn254.NewSuiteG2()
	s.SetDomainG1([]byte(DST))
	return s
}

func mustPubKey() kyber.Point {
	b, err := hex.DecodeString(PublicKeyHex)
	if err != nil {
		panic(err)
	}
	p := suite.G2().Point()
	if err := p.UnmarshalBinary(b); err != nil {
		panic(err)
	}
	return p
}

// Message returns the digest drand signs for an unchained round:
// keccak256(uint64_be(round)).
func Message(round uint64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], round)
	h := sha3.NewLegacyKeccak256()
	h.Write(b[:])
	return h.Sum(nil)
}

// Verify checks a beacon signature for round against the evmnet group key.
func Verify(round uint64, sig []byte) error {
	if len(sig) != SigSize {
		return fmt.Errorf("signature must be %d bytes, got %d", SigSize, len(sig))
	}
	return scheme.Verify(pubKey, Message(round), sig)
}

// Randomness derives the beacon randomness the way drand publishes it.
func Randomness(sig []byte) [32]byte {
	return sha256.Sum256(sig)
}

// HashToG1 hashes msg to G1 with the evmnet DST and returns the 64-byte
// uncompressed affine encoding (x|y, big-endian).
func HashToG1(msg []byte) []byte {
	p := suite.G1().Point().(kyber.HashablePoint).Hash(msg)
	b, err := p.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return b
}

// RoundAt returns the latest round published at or before unix time t.
func RoundAt(t int64) uint64 {
	if t < GenesisTime {
		return 0
	}
	return uint64((t-GenesisTime)/Period) + 1
}

// TimeOf returns the unix time at which round is published.
func TimeOf(round uint64) int64 {
	if round == 0 {
		return GenesisTime
	}
	return GenesisTime + int64(round-1)*Period
}
