package types

import (
	"sort"
	"testing"

	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EXPLORATORY: end-to-end tests covering a mixed-scheme validator set
// signing and verifying blocks. These confirm at the consensus layer
// that nothing assumes a single signing scheme; per-validator
// VerifyBytes is polymorphic and addresses do not collide across
// schemes.

// mixedValSet returns a validator set built from the supplied private
// keys (one validator per key, equal voting power) together with the
// matching PrivValidators, sorted by address.
func mixedValSet(t *testing.T, privKeys []crypto.PrivKey, votingPower int64) (*ValidatorSet, []PrivValidator) {
	t.Helper()
	require.NotEmpty(t, privKeys)

	valz := make([]*Validator, len(privKeys))
	privVals := make([]PrivValidator, len(privKeys))
	for i, pk := range privKeys {
		valz[i] = NewValidator(pk.PubKey(), votingPower)
		privVals[i] = NewMockPVWithPrivKey(pk)
	}
	// NewValidatorSet sorts validators by address internally; sort
	// privValidators the same way so MakeCommit's index assumptions hold.
	sort.Sort(PrivValidatorsByAddress(privVals))
	return NewValidatorSet(valz), privVals
}

func TestMixedScheme_AddressesDoNotCollide(t *testing.T) {
	t.Parallel()

	// Sanity check: ed25519 and secp256k1 derive 20-byte addresses via
	// disjoint hash chains (SHA256-trunc20 vs RIPEMD160(SHA256(.))).
	// Generating many of each should never collide.
	const n = 1000
	seen := make(map[crypto.Address]string, 2*n)

	for i := 0; i < n; i++ {
		ed := ed25519.GenPrivKey().PubKey().Address()
		if prev, ok := seen[ed]; ok {
			t.Fatalf("ed25519 address collision against %s: %v", prev, ed)
		}
		seen[ed] = "ed25519"

		sp := secp256k1.GenPrivKey().PubKey().Address()
		if prev, ok := seen[sp]; ok {
			t.Fatalf("cross-scheme address collision (ed25519 vs secp256k1): %s vs %v", prev, sp)
		}
		seen[sp] = "secp256k1"
	}
}

func TestMixedScheme_VerifyCommit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		factory func() []crypto.PrivKey
	}{
		{
			name: "all ed25519 (baseline)",
			factory: func() []crypto.PrivKey {
				return []crypto.PrivKey{
					ed25519.GenPrivKey(), ed25519.GenPrivKey(),
					ed25519.GenPrivKey(), ed25519.GenPrivKey(),
				}
			},
		},
		{
			name: "all secp256k1",
			factory: func() []crypto.PrivKey {
				return []crypto.PrivKey{
					secp256k1.GenPrivKey(), secp256k1.GenPrivKey(),
					secp256k1.GenPrivKey(), secp256k1.GenPrivKey(),
				}
			},
		},
		{
			name: "mixed 2 ed25519 + 2 secp256k1",
			factory: func() []crypto.PrivKey {
				return []crypto.PrivKey{
					ed25519.GenPrivKey(), secp256k1.GenPrivKey(),
					ed25519.GenPrivKey(), secp256k1.GenPrivKey(),
				}
			},
		},
		{
			name: "mixed 3 ed25519 + 1 secp256k1 (single HSM)",
			factory: func() []crypto.PrivKey {
				return []crypto.PrivKey{
					ed25519.GenPrivKey(), ed25519.GenPrivKey(),
					ed25519.GenPrivKey(), secp256k1.GenPrivKey(),
				}
			},
		},
		{
			name: "mixed 1 ed25519 + 3 secp256k1 (mostly HSM)",
			factory: func() []crypto.PrivKey {
				return []crypto.PrivKey{
					ed25519.GenPrivKey(),
					secp256k1.GenPrivKey(), secp256k1.GenPrivKey(), secp256k1.GenPrivKey(),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			valSet, privVals := mixedValSet(t, tt.factory(), 100)

			const (
				chainID = "mixed-scheme-chain"
				height  = int64(7)
				round   = 0
			)
			blockID := BlockID{Hash: []byte("block-hash-mixed-scheme")}

			voteSet := NewVoteSet(chainID, height, round, PrecommitType, valSet)
			commit, err := MakeCommit(blockID, height, round, voteSet, privVals)
			require.NoError(t, err)

			require.NoError(t, valSet.VerifyCommit(chainID, blockID, height, commit))
		})
	}
}

func TestMixedScheme_VerifyCommit_DetectsForgery(t *testing.T) {
	t.Parallel()

	// Build a 4-validator mixed set; commit signed by all 4; then mutate
	// one precommit's signature and confirm verification fails. This
	// guards against a subtle bug where a polymorphic VerifyBytes
	// accidentally treats a mismatched scheme's signature as valid.
	keys := []crypto.PrivKey{
		ed25519.GenPrivKey(), secp256k1.GenPrivKey(),
		ed25519.GenPrivKey(), secp256k1.GenPrivKey(),
	}
	valSet, privVals := mixedValSet(t, keys, 100)

	const (
		chainID = "forgery-chain"
		height  = int64(11)
	)
	blockID := BlockID{Hash: []byte("block-hash-forgery")}

	voteSet := NewVoteSet(chainID, height, 0, PrecommitType, valSet)
	commit, err := MakeCommit(blockID, height, 0, voteSet, privVals)
	require.NoError(t, err)
	require.NoError(t, valSet.VerifyCommit(chainID, blockID, height, commit))

	// Find a secp256k1 precommit and flip a signature byte.
	for i, pc := range commit.Precommits {
		_, val := valSet.GetByIndex(i)
		if _, ok := val.PubKey.(secp256k1.PubKeySecp256k1); !ok {
			continue
		}
		require.NotEmpty(t, pc.Signature)
		pc.Signature = append([]byte(nil), pc.Signature...)
		pc.Signature[0] ^= 0xff
		break
	}

	assert.Error(t, valSet.VerifyCommit(chainID, blockID, height, commit),
		"VerifyCommit must reject a forged secp256k1 signature")
}

func TestMixedScheme_VoteRoundTrip(t *testing.T) {
	t.Parallel()

	// Drive a single-vote round-trip for each scheme. Confirms PrivValidator
	// (the wrapper PrivVal interface used throughout consensus) handles
	// each scheme uniformly.
	cases := []struct {
		name string
		priv crypto.PrivKey
	}{
		{"ed25519", ed25519.GenPrivKey()},
		{"secp256k1", secp256k1.GenPrivKey()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			val := NewValidator(tc.priv.PubKey(), 1000)
			privVal := NewMockPVWithPrivKey(tc.priv)

			const (
				chainID = "single-validator-chain"
				height  = int64(3)
			)
			blockID := BlockID{Hash: []byte("block-hash-roundtrip")}
			vote := &Vote{
				ValidatorAddress: val.Address,
				ValidatorIndex:   0,
				Height:           height,
				Round:            0,
				Timestamp:        tmtime.Now(),
				Type:             PrecommitType,
				BlockID:          blockID,
			}
			require.NoError(t, privVal.SignVote(chainID, vote))

			// Inline verification using the same code path VerifyCommit uses.
			assert.True(t, val.PubKey.VerifyBytes(vote.SignBytes(chainID), vote.Signature))

			// Wrong-chain rejection (basic safety: changing the chainID
			// changes signBytes, so the signature must no longer verify).
			assert.False(t, val.PubKey.VerifyBytes(vote.SignBytes("other-chain"), vote.Signature))
		})
	}
}

func TestMixedScheme_Validators_AreDistinct(t *testing.T) {
	t.Parallel()

	// Build a mixed set and confirm GetByAddress finds each validator
	// regardless of scheme. Guards against a hypothetical map / hash
	// collision in the validator index.
	keys := []crypto.PrivKey{
		ed25519.GenPrivKey(), secp256k1.GenPrivKey(),
		ed25519.GenPrivKey(), secp256k1.GenPrivKey(),
	}
	valSet, _ := mixedValSet(t, keys, 100)

	for _, pk := range keys {
		addr := pk.PubKey().Address()
		idx, v := valSet.GetByAddress(addr)
		require.GreaterOrEqual(t, idx, 0, "validator missing for addr %v", addr)
		assert.Equal(t, addr, v.Address)
		assert.True(t, v.PubKey.Equals(pk.PubKey()))
	}
}
