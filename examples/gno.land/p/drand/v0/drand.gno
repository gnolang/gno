// Package drand verifies drand "evmnet" beacons on gno.land and turns them
// into randomness.
//
// evmnet is the drand network that signs on the BN254 curve (scheme
// bls-bn254-unchained-on-g1, one beacon every 3 seconds). gno.land ships
// BN254 as native precompiles, so a beacon is checked fully on-chain: no
// trusted relayer, no oracle committee. Whoever submits a beacon can only
// submit the real one.
//
// Verification: e(sig, G2) == e(H(m), pk), with m = keccak256(round) and H
// the RFC 9380 SVDW hash-to-curve, checked as one native pairing product.
package drand

import (
	"crypto/bn254"
	"crypto/keccak256"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
)

// evmnet network parameters, from https://api.drand.sh/<ChainHash>/info.
const (
	ChainHash   = "04f1e9062b8a81f848fded9c12306733282b2727ecced50032187751166ec8c3"
	SchemeID    = "bls-bn254-unchained-on-g1"
	GenesisTime = int64(1727521075)
	Period      = int64(3)
	SigSize     = 64
)

var (
	// Group public key on G2, encoded x_im|x_re|y_im|y_re (EIP-197 order).
	pubKey = mustHex("07e1d1d335df83fa98462005690372c643340060d205306a9aa8106b6bd0b3820557ec32c2ad488e4d4f6008f89a346f18492092ccc0d594610de2732c8b808f0095685ae3a85ba243747b1b2f426049010f6b73a0cf1d389351d5aaaa1047f6297d3a4f9749b33eb2d904c9d9ebf17224150ddd7abd7567a9bec6c74480ee0b")
	// Negated G2 generator, same encoding (y_im and y_re replaced by p - y).
	negG2Gen = mustHex("198e9393920d483a7260bfb731fb5d25f1aa493335a9e71297e485b7aef312c21800deef121f1e76426a00665e5c4479674322d4f75edadd46debd5cd992f6ed275dc4a288d1afb3cbb1ac09187524c7db36395df7be3b99e673b13a075a65ec1d9befcd05a5323e6da4d435f3b617cdb3af83285c2df711ef39c01571827f9d")
)

// RoundAt returns the latest round published at or before unix time t
// (0 before genesis).
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

// Message returns the digest drand signs for round: keccak256(uint64_be(round)).
func Message(round uint64) [32]byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], round)
	return keccak256.Sum256(b[:])
}

// Verify reports whether sig is the evmnet signature of round, checking
// e(sig, -G2) * e(H(m), pk) == 1. The pairing native rejects points off the
// curve and non-canonical coordinates.
func Verify(round uint64, sig []byte) bool {
	if round == 0 || len(sig) != SigSize {
		return false
	}
	msg := Message(round)
	in := make([]byte, 0, 2*bn254.PairingChunkSize)
	in = append(in, sig...)
	in = append(in, negG2Gen...)
	in = append(in, HashToG1(msg[:])...)
	in = append(in, pubKey...)
	out := bn254.PairingCheck(in)
	return len(out) == bn254.PairingOutputSize && out[31] == 1
}

// Randomness returns the beacon randomness for sig, sha256(sig), the same
// value drand publishes as "randomness".
func Randomness(sig []byte) [32]byte {
	return sha256.Sum256(sig)
}

// Derive returns sha256(beacon || salt). Use a distinct salt per consumer and
// per draw so that two users of the same round get independent values.
func Derive(beacon [32]byte, salt string) [32]byte {
	buf := make([]byte, 0, 32+len(salt))
	buf = append(buf, beacon[:]...)
	buf = append(buf, salt...)
	return sha256.Sum256(buf)
}

// NewRand returns a deterministic PRNG seeded from 32 bytes of randomness,
// for drawing several values (IntN, Perm, Shuffle) from one beacon.
func NewRand(seed [32]byte) *rand.Rand {
	be := binary.BigEndian
	return rand.New(rand.NewPCG(be.Uint64(seed[0:8]), be.Uint64(seed[8:16])))
}

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func mustFe(s string) fe {
	e, ok := feFromCanonical(mustHex(s))
	if !ok {
		panic("drand: constant not reduced")
	}
	return e
}
