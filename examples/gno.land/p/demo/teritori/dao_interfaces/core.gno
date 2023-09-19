package dao_interfaces

type ActivableProposalModule struct {
	Enabled bool
	Module  IProposalModule
}

type IDAOCore interface {
	Render(path string) string

	VotingModule() IVotingModule
	ProposalModules() []ActivableProposalModule
	ActiveProposalModuleCount() int
	Registry() *MessagesRegistry

	UpdateVotingModule(newVotingModule IVotingModule)
	UpdateProposalModules(toAdd []IProposalModule, toDisable []int)
}
