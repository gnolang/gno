package cstypes

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
)

type ConsensusEvent interface {
	types.TMEvent
	GetHRS() HRS
}

func (EventNewRoundStep) AssertEvent()     {}
func (EventNewValidBlock) AssertEvent()    {}
func (EventNewRound) AssertEvent()         {}
func (EventCompleteProposal) AssertEvent() {}
func (EventTimeoutPropose) AssertEvent()   {}
func (EventTimeoutWait) AssertEvent()      {}
func (EventPolka) AssertEvent()            {}
func (EventLock) AssertEvent()             {}
func (EventUnlock) AssertEvent()           {}
func (EventRelock) AssertEvent()           {}

var (
	_ ConsensusEvent = EventNewRoundStep{}
	_ ConsensusEvent = EventNewValidBlock{}
	_ ConsensusEvent = EventNewRound{}
	_ ConsensusEvent = EventCompleteProposal{}
	_ ConsensusEvent = EventTimeoutPropose{}
	_ ConsensusEvent = EventTimeoutWait{}
	_ ConsensusEvent = EventPolka{}
	_ ConsensusEvent = EventLock{}
	_ ConsensusEvent = EventUnlock{}
	_ ConsensusEvent = EventRelock{}
)

type EventNewRoundStep struct {
	HRS `json:"hrs"` // embed for "GetHRS()"

	SecondsSinceStartTime int
	LastCommitRound       int
}

func (ev EventNewRoundStep) String() string {
	return fmt.Sprintf("EventNewRoundStep{%v}", ev.HRS)
}

type EventNewValidBlock struct {
	HRS `json:"hrs"`

	BlockPartsHeader types.PartSetHeader `json:"block_parts_header"`
	BlockParts       *bitarray.BitArray  `json:"block_parts"`
	IsCommit         bool                `json:"is_commit"`
}

func (ev EventNewValidBlock) String() string {
	return fmt.Sprintf("EventNewValidBlock{%v}", ev.HRS)
}

type EventNewRound struct {
	HRS `json:"hrs"`

	Proposer      types.Validator `json:"proposer"`
	ProposerIndex int             `json:"proposer_index"`
}

func (ev EventNewRound) String() string {
	return fmt.Sprintf("EventNewRound{%v}", ev.HRS)
}

type EventCompleteProposal struct {
	HRS `json:"hrs"`

	BlockID types.BlockID `json:"block_id"`
}

func (ev EventCompleteProposal) String() string {
	return fmt.Sprintf("EventCompleteProposal{%v}", ev.HRS)
}

type EventTimeoutPropose struct {
	HRS `json:"hrs"`
}

func (ev EventTimeoutPropose) String() string {
	return fmt.Sprintf("EventTimeoutPropose{%v}", ev.HRS)
}

type EventTimeoutWait struct {
	HRS `json:"hrs"`
}

func (ev EventTimeoutWait) String() string {
	return fmt.Sprintf("EventTimeoutWait{%v}", ev.HRS)
}

type EventPolka struct {
	HRS `json:"hrs"`
}

func (ev EventPolka) String() string {
	return fmt.Sprintf("EventPolka{%v}", ev.HRS)
}

type EventLock struct {
	HRS `json:"hrs"`
}

func (ev EventLock) String() string {
	return fmt.Sprintf("EventLock{%v}", ev.HRS)
}

type EventUnlock struct {
	HRS `json:"hrs"`
}

func (ev EventUnlock) String() string {
	return fmt.Sprintf("EventUnlock{%v}", ev.HRS)
}

type EventRelock struct {
	HRS `json:"hrs"`
}

func (ev EventRelock) String() string {
	return fmt.Sprintf("EventRelock{%v}", ev.HRS)
}
