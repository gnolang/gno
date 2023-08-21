package core

import (
	"std"
	"strings"

	dao_interfaces "gno.land/p/demo/daodao/interfaces_v3"
	"gno.land/p/demo/markdown_utils"
)

// TODO: add wrapper message handler to handle multiple proposal modules messages

type IDAOCore interface {
	AddProposalModule(proposalMod dao_interfaces.IProposalModule)

	VotingModule() dao_interfaces.IVotingModule
	ProposalModules() []dao_interfaces.IProposalModule

	Render(path string) string
}

type daoCore struct {
	IDAOCore

	votingModule    dao_interfaces.IVotingModule
	proposalModules []dao_interfaces.IProposalModule
}

func NewDAOCore(
	votingModule dao_interfaces.IVotingModule,
	proposalModules []dao_interfaces.IProposalModule,
) IDAOCore {
	return &daoCore{
		votingModule:    votingModule,
		proposalModules: proposalModules,
	}
}

func (d *daoCore) VotingModule() dao_interfaces.IVotingModule {
	return d.votingModule
}

func (d *daoCore) ProposalModules() []dao_interfaces.IProposalModule {
	return d.proposalModules
}

func (d *daoCore) AddProposalModule(proposalMod dao_interfaces.IProposalModule) {
	d.proposalModules = append(d.proposalModules, proposalMod)
}

func (d *daoCore) Render(path string) string {
	s := "# DAO Core\n"
	s += "This is a port of [DA0-DA0 contracts](https://github.com/DA0-DA0/dao-contracts)\n"
	s += markdown_utils.Indent(d.votingModule.Render(path)) + "\n"
	for _, propMod := range d.proposalModules {
		s += markdown_utils.Indent(propMod.Render(path)) + "\n"
	}
	return s
}

func GetProposalModule(core IDAOCore, moduleIndex int) dao_interfaces.IProposalModule {
	if moduleIndex < 0 {
		panic("Module index must be >= 0")
	}
	mods := core.ProposalModules()
	if moduleIndex >= len(mods) {
		panic("invalid module index")
	}
	return mods[moduleIndex]
}
