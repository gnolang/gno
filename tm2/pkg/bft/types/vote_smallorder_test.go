package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// TestSmallOrderValidatorCannotSign covers the consensus path. A validator is
// identified by the address derived from its registered pubkey, so a
// small-order key cannot be substituted for an honest validator. It can,
// however, be admitted as a validator in its own right (through a governance
// proposal, or a buggy operator). Before the VerifyBytes gate such a validator
// could be signed for by anyone: the forged (R = identity, S = 0) precommit
// verified, and with enough power it also satisfied VerifyCommit. Now it
// produces no valid vote at all.
func TestSmallOrderValidatorCannotSign(t *testing.T) {
	var pk ed25519.PubKeyEd25519
	pk[0] = 1 // identity point
	val := NewValidator(pk, 1000)
	valset := NewValidatorSet([]*Validator{val})

	chainID := "small-order-chain"
	blockID := BlockID{Hash: []byte("block")}
	height := int64(5)

	forged := make([]byte, ed25519.SignatureSize)
	forged[0] = 1 // R = identity, S = 0

	vote := &Vote{
		ValidatorAddress: val.Address,
		ValidatorIndex:   0,
		Height:           height,
		Round:            0,
		Timestamp:        tmtime.Now(),
		Type:             PrecommitType,
		BlockID:          blockID,
		Signature:        forged,
	}

	// The address matches by construction, so this is the signature gate.
	require.ErrorIs(t, vote.Verify(chainID, pk), ErrVoteInvalidSignature)

	commit := NewCommit(blockID, []*CommitSig{vote.CommitSig()})
	require.Error(t, valset.VerifyCommit(chainID, blockID, height, commit))
}
