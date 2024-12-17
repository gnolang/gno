package core

import (
	"std"
	"testing"

	dao_interfaces "gno.land/p/teritori/dao_interfaces"
)

type votingModule struct {
	core dao_interfaces.IDAOCore
}

func votingModuleFactory(core dao_interfaces.IDAOCore) dao_interfaces.IVotingModule {
	return &votingModule{core: core}
}

func (vm *votingModule) Core() dao_interfaces.IDAOCore {
	return vm.core
}

func (vm *votingModule) Info() dao_interfaces.ModuleInfo {
	return dao_interfaces.ModuleInfo{
		Kind:    "TestVoting",
		Version: "21.42",
	}
}

func (vm *votingModule) ConfigJSON() string {
	return "{}"
}

func (vm *votingModule) GetMembersJSON(start, end string, limit uint64, height int64) string {
	return "[]"
}

func (vm *votingModule) Render(path string) string {
	return "# Test Voting Module"
}

func (vm *votingModule) VotingPowerAtHeight(address std.Address, height int64, resources []string) uint64 {
	return 0
}

func (vm *votingModule) TotalPowerAtHeight(height int64, resources []string) uint64 {
	return 0
}

type proposalModule struct {
	core dao_interfaces.IDAOCore
}

func proposalModuleFactory(core dao_interfaces.IDAOCore) dao_interfaces.IProposalModule {
	return &proposalModule{core: core}
}

func (pm *proposalModule) Core() dao_interfaces.IDAOCore {
	return pm.core
}

func (pm *proposalModule) Info() dao_interfaces.ModuleInfo {
	return dao_interfaces.ModuleInfo{
		Kind:    "TestProposal",
		Version: "42.21",
	}
}

func (pm *proposalModule) ConfigJSON() string {
	return "{}"
}

func (pm *proposalModule) VoteJSON(proposalID int, voteJSON string) {
	panic("not implemented")
}

func (pm *proposalModule) Render(path string) string {
	return "# Test Proposal Module"
}

func (pm *proposalModule) Execute(proposalId int) {
	panic("not implemented")
}

func (pm *proposalModule) ProposeJSON(proposalJSON string) int {
	panic("not implemented")
}

func (pm *proposalModule) ProposalsJSON(limit int, startAfter string, reverse bool) string {
	panic("not implemented")
}

func (pm *proposalModule) ProposalJSON(proposalID int) string {
	panic("not implemented")
}

func TestDAOCore(t *testing.T) {
	var testValue string
	handler := dao_interfaces.NewCopyMessageHandler(&testValue)
	handlerFactory := func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
		return handler
	}

	core := NewDAOCore(votingModuleFactory, []dao_interfaces.ProposalModuleFactory{proposalModuleFactory}, []dao_interfaces.MessageHandlerFactory{handlerFactory})
	if core == nil {
		t.Fatal("core is nil")
	}

	if core.ActiveProposalModuleCount() != 1 {
		t.Fatal("expected 1 active proposal module")
	}

	votingMod := core.VotingModule()
	if votingMod == nil {
		t.Fatal("voting module is nil")
	}

	if votingMod.Info().Kind != "TestVoting" {
		t.Fatal("voting module has wrong kind")
	}

	propMods := core.ProposalModules()
	if len(propMods) != 1 {
		t.Fatal("expected 1 proposal module")
	}

	propMod := propMods[0]
	if !propMod.Enabled {
		t.Fatal("proposal module is not enabled")
	}

	if propMod.Module == nil {
		t.Fatal("proposal module is nil")
	}

	if propMod.Module.Info().Kind != "TestProposal" {
		t.Fatal("proposal module has wrong kind")
	}

	registry := core.Registry()
	if registry == nil {
		t.Fatal("registry is nil")
	}

	msg := &dao_interfaces.CopyMessage{Value: "test"}
	registry.Execute(msg)
	if testValue != "test" {
		t.Errorf("expected testValue to be 'test', got '%s'", testValue)
	}

	newProposalModule := &proposalModule{core: core}
	updatePropModsMsg := &UpdateProposalModulesExecutableMessage{
		ToAdd:     []dao_interfaces.IProposalModule{newProposalModule},
		ToDisable: []int{0},
	}
	registry.Execute(updatePropModsMsg)

	if core.ActiveProposalModuleCount() != 1 {
		t.Fatal("expected 1 active proposal module")
	}

	propMods = core.ProposalModules()
	if len(propMods) != 2 {
		t.Fatal("expected 2 proposal modules")
	}

	propMod = propMods[0]
	if propMod.Enabled {
		t.Errorf("old proposal module is still enabled")
	}

	propMod = propMods[1]
	if !propMod.Enabled {
		t.Errorf("new proposal module is not enabled")
	}

	if propMod.Module != newProposalModule {
		t.Errorf("new proposal module is not the same as the one added")
	}

	newVotingModule := &votingModule{core: core}
	updateVotingModMsg := &UpdateVotingModuleExecutableMessage{
		Module: newVotingModule,
	}

	registry.Execute(updateVotingModMsg)

	votingMod = core.VotingModule()
	if votingMod != newVotingModule {
		t.Errorf("new voting module is not the same as the one added")
	}
}
