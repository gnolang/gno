package types

import (
	"errors"
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

const (
	// MaxVoteBytes is a maximum vote size (including amino overhead).
	MaxVoteBytes int    = 247
	nilVoteStr   string = "nil-Vote"
)

var (
	ErrVoteUnexpectedStep            = errors.New("unexpected step")
	ErrVoteInvalidValidatorIndex     = errors.New("invalid validator index")
	ErrVoteInvalidValidatorAddress   = errors.New("invalid validator address")
	ErrVoteInvalidSignature          = errors.New("invalid signature")
	ErrVoteNonDeterministicSignature = errors.New("non-deterministic signature")
	ErrVoteNil                       = errors.New("nil vote")
)

type VoteConflictingVotesError struct {
	*DuplicateVoteEvidence
}

func (err *VoteConflictingVotesError) Error() string {
	return fmt.Sprintf("Conflicting votes from validator %v", err.PubKey.Address())
}

func NewConflictingVoteError(val *Validator, voteA, voteB *Vote) *VoteConflictingVotesError {
	return &VoteConflictingVotesError{
		&DuplicateVoteEvidence{
			PubKey: val.PubKey,
			VoteA:  voteA,
			VoteB:  voteB,
		},
	}
}

// Address is hex bytes.
type Address = crypto.Address

// Vote represents a prevote, precommit, or commit vote from validators for
// consensus.
type Vote struct {
	Type             SignedMsgType `json:"type"`
	Height           int64         `json:"height"`
	Round            int           `json:"round"`
	BlockID          BlockID       `json:"block_id"` // zero if vote is nil.
	Timestamp        time.Time     `json:"timestamp"`
	ValidatorAddress Address       `json:"validator_address"`
	ValidatorIndex   int           `json:"validator_index"`
	Signature        []byte        `json:"signature"`
}

// CommitSig converts the Vote to a CommitSig.
// If the Vote is nil, the CommitSig will be nil.
func (vote *Vote) CommitSig() *CommitSig {
	if vote == nil {
		return nil
	}
	cs := CommitSig(*vote)
	return &cs
}

func (vote *Vote) SignBytes(chainID string) []byte {
	bz, err := amino.MarshalSized(CanonicalizeVote(chainID, vote))
	if err != nil {
		panic(err)
	}
	return bz
}

func (vote *Vote) Copy() *Vote {
	voteCopy := *vote
	return &voteCopy
}

func (vote *Vote) String() string {
	if vote == nil {
		return nilVoteStr
	}
	var typeString string
	switch vote.Type {
	case PrevoteType:
		typeString = "Prevote"
	case PrecommitType:
		typeString = "Precommit"
	default:
		panic("Unknown vote type")
	}

	return fmt.Sprintf("Vote{%v:%X %v/%02d/%v(%v) %X %X @ %s}",
		vote.ValidatorIndex,
		fingerprint(vote.ValidatorAddress[:]),
		vote.Height,
		vote.Round,
		vote.Type,
		typeString,
		fingerprint(vote.BlockID.Hash),
		fingerprint(vote.Signature),
		CanonicalTime(vote.Timestamp),
	)
}

func (vote *Vote) Verify(chainID string, pubKey crypto.PubKey) error {
	if pubKey.Address() != vote.ValidatorAddress {
		return ErrVoteInvalidValidatorAddress
	}

	if !pubKey.VerifyBytes(vote.SignBytes(chainID), vote.Signature) {
		return ErrVoteInvalidSignature
	}
	return nil
}

// ValidateBasic performs basic validation.
func (vote *Vote) ValidateBasic() error {
	if !IsVoteTypeValid(vote.Type) {
		return errors.New("invalid Type")
	}
	if vote.Height < 0 {
		return errors.New("negative Height")
	}
	if vote.Round < 0 {
		return errors.New("negative Round")
	}

	// NOTE: Timestamp validation is subtle and handled elsewhere.

	if err := vote.BlockID.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong BlockID: %w", err)
	}
	// BlockID.ValidateBasic would not err if we for instance have an empty hash but a
	// non-empty PartsSetHeader:
	if !vote.BlockID.IsZero() && !vote.BlockID.IsComplete() {
		return fmt.Errorf("BlockID must be either empty or complete, got: %v", vote.BlockID)
	}
	if len(vote.ValidatorAddress) != crypto.AddressSize {
		return fmt.Errorf("expected ValidatorAddress size to be %d bytes, got %d bytes",
			crypto.AddressSize,
			len(vote.ValidatorAddress),
		)
	}
	if vote.ValidatorAddress.IsZero() {
		return fmt.Errorf("expected ValidatorAddress to be non-zero")
	}
	if vote.ValidatorIndex < 0 {
		return errors.New("negative ValidatorIndex")
	}
	if len(vote.Signature) == 0 {
		return errors.New("signature is missing")
	}
	if len(vote.Signature) > MaxSignatureSize {
		return fmt.Errorf("signature is too big (max: %d)", MaxSignatureSize)
	}
	return nil
}
