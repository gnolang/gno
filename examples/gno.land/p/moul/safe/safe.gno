// Package safe is inspired by Gnosis Safe
package safe

import (
	"std"

	"gno.land/p/demo/avl"
)

//  Gnosis safe style using this with threshold
// : affMember + update threshold at the same time

type Safe interface {
	Name() string
	Members() []std.Address
	Threshold() uint
	Size() uint
	NextProposal() Proposal
	UpcomingProposals(page uint) (props []Proposal, total uint)
	PreviousProposals(page uint) (props []Proposal, total uint)
	// balances (native, grc20, grc721)

	ProposeAddingMember(addr std.Address, newThreshold uint) (prop Proposal, err error)
	ProposeRemovingMember(addr std.Address, newThreshold uint) (prop Proposal, err error)
	ProposeUpdatingThreshold(newThreshold uint) (prop Proposal, err error)
	// ProposeClosure(cl func()) (prop Proposal, err error)
	// ProposeBankXXX()
	// ProposeGRC20XXX()
	// ProposeGRC721XXX()

	GetProposal(id uint) Proposal
	ApproveProposal(p Proposal) error
	ExecuteProposal(p Proposal) error
	CancelProposal(p Proposal) error // XXX: should be ProposeToCancel?

	Render(path string) string
}

type Proposal interface {
	ID() uint
	Active() bool
	Executed() bool
	Voters() []std.Address
}

func NewSafe() Safe {
	caller := std.PrevRealm().Addr()
	return safe{
		members:   []std.Address{caller},
		threshold: 1,
		prevProps: []Proposal{},
		nextProps: []Proposal{},
	}
}

type safe struct {
	name      string
	members   avl.Tree // id -> std.Address
	threshold uint
	prevProps avl.tree // pid -> Proposal
	nextProps avl.Tree // pid -> Proposal
}

func (s safe) String() string { panic("not implemented") }

func (s safe) Members() []std.Address { return s.members }
func (s safe) Threshold() uint        { return s.threshold }
func (s safe) Size() uint             { return len(s.members) }
func (s *safe) AddMember(addr std.Address, newThreshold uint) error {
}

type proposal struct{}

func (p proposal) String() string { panic("not implemented") }
