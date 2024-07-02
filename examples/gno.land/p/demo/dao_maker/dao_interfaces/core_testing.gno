package dao_interfaces

type dummyCore struct{}

func NewDummyCore() IDAOCore {
	return &dummyCore{}
}

func (d *dummyCore) Render(path string) string {
	panic("not implemented")
}

func (d *dummyCore) VotingModule() IVotingModule {
	panic("not implemented")
}

func (d *dummyCore) ProposalModules() []ActivableProposalModule {
	panic("not implemented")
}

func (d *dummyCore) ActiveProposalModuleCount() int {
	panic("not implemented")
}

func (d *dummyCore) Registry() *MessagesRegistry {
	panic("not implemented")
}

func (d *dummyCore) UpdateVotingModule(newVotingModule IVotingModule) {
	panic("not implemented")
}

func (d *dummyCore) UpdateProposalModules(toAdd []IProposalModule, toDisable []int) {
	panic("not implemented")
}
