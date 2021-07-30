package types

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/bitarray"
)

type ConsensusEvent interface {
	types.TMEvent
	GetHRS() HRS
}

func (_ EventNewRoundStep) AssertEvent()     {}
func (_ EventNewValidBlock) AssertEvent()    {}
func (_ EventNewRound) AssertEvent()         {}
func (_ EventCompleteProposal) AssertEvent() {}
func (_ EventTimeoutPropose) AssertEvent()   {}
func (_ EventTimeoutWait) AssertEvent()      {}
func (_ EventPolka) AssertEvent()            {}
func (_ EventLock) AssertEvent()             {}
func (_ EventUnlock) AssertEvent()           {}
func (_ EventRelock) AssertEvent()           {}

var _ ConsensusEvent = EventNewRoundStep{}
var _ ConsensusEvent = EventNewValidBlock{}
var _ ConsensusEvent = EventNewRound{}
var _ ConsensusEvent = EventCompleteProposal{}
var _ ConsensusEvent = EventTimeoutPropose{}
var _ ConsensusEvent = EventTimeoutWait{}
var _ ConsensusEvent = EventPolka{}
var _ ConsensusEvent = EventLock{}
var _ ConsensusEvent = EventUnlock{}
var _ ConsensusEvent = EventRelock{}

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
