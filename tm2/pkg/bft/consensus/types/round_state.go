package cstypes

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

//-----------------------------------------------------------------------------
// RoundStepType enum type

// RoundStepType enumerates the state of the consensus state machine
type RoundStepType uint8 // These must be numeric, ordered.

// RoundStepType
const (
	RoundStepInvalid       = RoundStepType(0x00) // Invalid
	RoundStepNewHeight     = RoundStepType(0x01) // Wait til CommitTime + timeoutCommit
	RoundStepNewRound      = RoundStepType(0x02) // Setup new round and go to RoundStepPropose
	RoundStepPropose       = RoundStepType(0x03) // Did propose, gossip proposal
	RoundStepPrevote       = RoundStepType(0x04) // Did prevote, gossip prevotes
	RoundStepPrevoteWait   = RoundStepType(0x05) // Did receive any +2/3 prevotes, start timeout
	RoundStepPrecommit     = RoundStepType(0x06) // Did precommit, gossip precommits
	RoundStepPrecommitWait = RoundStepType(0x07) // Did receive any +2/3 precommits, start timeout
	RoundStepCommit        = RoundStepType(0x08) // Entered commit state machine
	// NOTE: RoundStepNewHeight acts as RoundStepCommitWait.

	// NOTE: Update IsValid method if you change this!
)

// IsValid returns true if the step is valid, false if unknown/undefined.
func (rs RoundStepType) IsValid() bool {
	return RoundStepNewHeight <= rs && rs <= RoundStepCommit
}

// String returns a string
func (rs RoundStepType) String() string {
	switch rs {
	case RoundStepInvalid:
		return "RoundStepInvalid"
	case RoundStepNewHeight:
		return "RoundStepNewHeight"
	case RoundStepNewRound:
		return "RoundStepNewRound"
	case RoundStepPropose:
		return "RoundStepPropose"
	case RoundStepPrevote:
		return "RoundStepPrevote"
	case RoundStepPrevoteWait:
		return "RoundStepPrevoteWait"
	case RoundStepPrecommit:
		return "RoundStepPrecommit"
	case RoundStepPrecommitWait:
		return "RoundStepPrecommitWait"
	case RoundStepCommit:
		return "RoundStepCommit"
	default:
		return "RoundStepUnknown" // Cannot panic.
	}
}

//-----------------------------------------------------------------------------

// RoundState defines the internal consensus state.
// NOTE: Not thread safe. Should only be manipulated by functions downstream
// of the cs.receiveRoutine
type RoundState struct {
	// TODO replace w/ HRS
	Height                    int64               `json:"height"` // Height we are working on
	Round                     int                 `json:"round"`
	Step                      RoundStepType       `json:"step"`
	StartTime                 time.Time           `json:"start_time"`
	CommitTime                time.Time           `json:"commit_time"` // Subjective time when +2/3 precommits for Block at Round were found
	Validators                *types.ValidatorSet `json:"validators"`
	Proposal                  *types.Proposal     `json:"proposal"`
	ProposalBlock             *types.Block        `json:"proposal_block"`
	ProposalBlockParts        *types.PartSet      `json:"proposal_block_parts"`
	LockedRound               int                 `json:"locked_round"`
	LockedBlock               *types.Block        `json:"locked_block"`
	LockedBlockParts          *types.PartSet      `json:"locked_block_parts"`
	ValidRound                int                 `json:"valid_round"`       // Last known round with POL for non-nil valid block.
	ValidBlock                *types.Block        `json:"valid_block"`       // Last known block of POL mentioned above.
	ValidBlockParts           *types.PartSet      `json:"valid_block_parts"` // Last known block parts of POL metnioned above.
	Votes                     *HeightVoteSet      `json:"votes"`
	CommitRound               int                 `json:"commit_round"` //
	LastCommit                *types.VoteSet      `json:"last_commit"`  // Last precommits at Height-1
	LastValidators            *types.ValidatorSet `json:"last_validators"`
	TriggeredTimeoutPrecommit bool                `json:"triggered_timeout_precommit"`
}

type HRS struct {
	Height int64         `json:"height"`
	Round  int           `json:"round"`
	Step   RoundStepType `json:"step"`
}

func (hrs HRS) Compare(other HRS) int {
	if hrs.Height < other.Height {
		return -1
	} else if hrs.Height == other.Height {
		if hrs.Round < other.Round {
			return -1
		} else if hrs.Round == other.Round {
			if hrs.Step < other.Step {
				return -1
			} else if hrs.Step == other.Step {
				return 0
			}
		}
	}
	return 1
}

func (hrs HRS) IsHRSZero() bool {
	return hrs == HRS{}
}

func (hrs HRS) GetHRS() HRS {
	return hrs
}

func (hrs HRS) String() string {
	return fmt.Sprintf("%d/%d/%d", hrs.Height, hrs.Round, hrs.Step)
}

func (rs *RoundState) GetHRS() HRS {
	if rs == nil {
		return HRS{0, 0, RoundStepInvalid}
	}
	return HRS{rs.Height, rs.Round, rs.Step}
}

// Compressed version of the RoundState for use in RPC
type RoundStateSimple struct {
	HeightRoundStep   string         `json:"height/round/step"`
	StartTime         time.Time      `json:"start_time"`
	ProposalBlockHash []byte         `json:"proposal_block_hash"`
	LockedBlockHash   []byte         `json:"locked_block_hash"`
	ValidBlockHash    []byte         `json:"valid_block_hash"`
	Votes             *HeightVoteSet `json:"height_vote_set"`
}

// Compress the RoundState to RoundStateSimple
func (rs *RoundState) RoundStateSimple() RoundStateSimple {
	return RoundStateSimple{
		HeightRoundStep:   rs.GetHRS().String(),
		StartTime:         rs.StartTime,
		ProposalBlockHash: rs.ProposalBlock.Hash(),
		LockedBlockHash:   rs.LockedBlock.Hash(),
		ValidBlockHash:    rs.ValidBlock.Hash(),
		Votes:             rs.Votes,
	}
}

func (rs *RoundState) EventNewRoundStep() EventNewRoundStep {
	return EventNewRoundStep{
		HRS:                   rs.GetHRS(),
		SecondsSinceStartTime: int(time.Since(rs.StartTime).Seconds()),
		LastCommitRound:       rs.LastCommit.Round(),
	}
}

// EventNewRound returns the RoundState with proposer information as an event.
func (rs *RoundState) EventNewRound() EventNewRound {
	proposer := rs.Validators.GetProposer()
	proposerIdx, _ := rs.Validators.GetByAddress(proposer.Address)

	return EventNewRound{
		HRS:           rs.GetHRS(),
		Proposer:      *proposer.Copy(),
		ProposerIndex: proposerIdx,
	}
}

// EventCompleteProposal returns information about a proposed block as an event.
func (rs *RoundState) EventCompleteProposal() EventCompleteProposal {
	// We must construct BlockID from ProposalBlock and ProposalBlockParts
	// cs.Proposal is not guaranteed to be set when this function is called
	blockId := types.BlockID{
		Hash:        rs.ProposalBlock.Hash(),
		PartsHeader: rs.ProposalBlockParts.Header(),
	}

	return EventCompleteProposal{
		HRS:     rs.GetHRS(),
		BlockID: blockId,
	}
}

func (rs *RoundState) EventNewValidBlock() EventNewValidBlock {
	return EventNewValidBlock{
		HRS:              rs.GetHRS(),
		BlockPartsHeader: rs.ProposalBlockParts.Header(),
		BlockParts:       rs.ProposalBlockParts.BitArray(),
		IsCommit:         rs.Step == RoundStepCommit,
	}
}

// String returns a string
func (rs *RoundState) String() string {
	return rs.StringIndented("")
}

// StringIndented returns a string
func (rs *RoundState) StringIndented(indent string) string {
	return fmt.Sprintf(`RoundState{
%s  H:%v R:%v S:%v
%s  StartTime:     %v
%s  CommitTime:    %v
%s  Validators:    %v
%s  Proposal:      %v
%s  ProposalBlock: %v %v
%s  LockedRound:   %v
%s  LockedBlock:   %v %v
%s  ValidRound:   %v
%s  ValidBlock:   %v %v
%s  Votes:         %v
%s  LastCommit:    %v
%s  LastValidators:%v
%s}`,
		indent, rs.Height, rs.Round, rs.Step,
		indent, rs.StartTime,
		indent, rs.CommitTime,
		indent, rs.Validators.StringIndented(indent+"  "),
		indent, rs.Proposal,
		indent, rs.ProposalBlockParts.StringShort(), rs.ProposalBlock.StringShort(),
		indent, rs.LockedRound,
		indent, rs.LockedBlockParts.StringShort(), rs.LockedBlock.StringShort(),
		indent, rs.ValidRound,
		indent, rs.ValidBlockParts.StringShort(), rs.ValidBlock.StringShort(),
		indent, rs.Votes.StringIndented(indent+"  "),
		indent, rs.LastCommit.StringShort(),
		indent, rs.LastValidators.StringIndented(indent+"  "),
		indent)
}

// StringShort returns a string
func (rs *RoundState) StringShort() string {
	return fmt.Sprintf(`RoundState{H:%v R:%v S:%v ST:%v}`,
		rs.Height, rs.Round, rs.Step, rs.StartTime)
}
