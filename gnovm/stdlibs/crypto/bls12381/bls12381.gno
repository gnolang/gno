// Package bls12381 exposes BLS signature verification over the BLS12-381
// curve as a native binding, with the parameters Ethereum's consensus layer
// uses: the IETF proof-of-possession ciphersuite
// (BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_) with public keys in G1 and
// signatures in G2, both in compressed (ZCash) encoding.
//
// The proof-of-possession ciphersuite is only secure against rogue-key
// attacks when every public key has passed PopVerify. This package does not
// provide PopVerify. Every public key passed to FastAggregateVerify or
// FastAggregateVerifyExcluding MUST have had its proof of possession
// verified, or come from a source that guarantees it: Ethereum sync
// committee keys qualify, since each passed PopVerify at deposit. Keys
// supplied by a relayer or any other untrusted party do not, and a realm
// aggregating them is forgeable however carefully it checks the rest.
package bls12381

const (
	// PubKeySize is the length of a compressed G1 public key.
	PubKeySize = 48
	// SignatureSize is the length of a compressed G2 signature.
	SignatureSize = 96
)

// FastAggregateVerify reports whether sig is a valid aggregate signature of
// msg by all of pubkeys, as specified by FastAggregateVerify in the IETF BLS
// signature draft. pubkeys is the concatenation of at least one PubKeySize
// public key, each of which MUST have passed proof-of-possession
// verification (see the package doc). Malformed keys or signatures, points
// outside the prime-order subgroup, an identity key and keys summing to the
// identity make it return false.
func FastAggregateVerify(pubkeys, msg, sig []byte) bool {
	// Reject wrong lengths before the native call so a malformed input is
	// not copied into Go memory first; see adr/native_input_bounds.md.
	if len(pubkeys) == 0 || len(pubkeys)%PubKeySize != 0 || len(sig) != SignatureSize {
		return false
	}
	return fastAggregateVerify(pubkeys, msg, sig)
}

// FastAggregateVerifyExcluding reports whether sig is a valid signature of
// msg by the members of a committee that are not in excluded, given the
// committee's aggregate public key: it is FastAggregateVerify over aggregate
// minus excluded. Use it only when aggregate is known to be the sum of the
// committee's keys (an Ethereum SyncCommittee commits to it); it then costs
// one key decoding per non-signer instead of one per signer.
//
// The caller MUST have verified that every key in excluded is a member of
// the committee behind aggregate, for example by a Merkle proof of each
// against the committee's pubkeys root, and every member MUST have passed
// proof-of-possession verification (see the package doc). This function
// cannot check membership: an attacker holding a key pk_a chooses excluded
// as e1 and aggregate - pk_a - e1, both valid subgroup points, so that
// aggregate minus excluded is exactly pk_a and the attacker's own signature
// verifies against the trusted aggregate for a message the committee never
// signed.
//
// aggregate is a PubKeySize key, excluded the concatenation of zero or more
// PubKeySize keys. Malformed or out-of-subgroup points, and a committee in
// which nobody signed, make it return false.
func FastAggregateVerifyExcluding(aggregate, excluded, msg, sig []byte) bool {
	// Reject wrong lengths before the native call; see FastAggregateVerify.
	if len(aggregate) != PubKeySize || len(excluded)%PubKeySize != 0 || len(sig) != SignatureSize {
		return false
	}
	return fastAggregateVerifyExcluding(aggregate, excluded, msg, sig)
}

func fastAggregateVerify(pubkeys, msg, sig []byte) bool // injected

func fastAggregateVerifyExcluding(aggregate, excluded, msg, sig []byte) bool // injected
