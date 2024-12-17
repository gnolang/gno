package core

import (
	"std"
	"strconv"
	"strings"

	dao_interfaces "gno.land/p/demo/dao_maker/dao_interfaces"
)

// TODO: add wrapper message handler to handle multiple proposal modules messages

type daoCore struct {
	dao_interfaces.IDAOCore

	votingModule              dao_interfaces.IVotingModule
	proposalModules           []dao_interfaces.ActivableProposalModule
	activeProposalModuleCount int
	realm                     std.Realm
	registry                  *dao_interfaces.MessagesRegistry
}

func NewDAOCore(
	votingModuleFactory dao_interfaces.VotingModuleFactory,
	proposalModulesFactories []dao_interfaces.ProposalModuleFactory,
	messageHandlersFactories []dao_interfaces.MessageHandlerFactory,
) dao_interfaces.IDAOCore {
	if votingModuleFactory == nil {
		panic("Missing voting module factory")
	}

	if len(proposalModulesFactories) == 0 {
		panic("No proposal modules factories")
	}

	core := &daoCore{
		realm:                     std.CurrentRealm(),
		activeProposalModuleCount: len(proposalModulesFactories),
		registry:                  dao_interfaces.NewMessagesRegistry(),
		proposalModules:           make([]dao_interfaces.ActivableProposalModule, len(proposalModulesFactories)),
	}

	core.votingModule = votingModuleFactory(core)
	if core.votingModule == nil {
		panic("voting module factory returned nil")
	}

	for i, modFactory := range proposalModulesFactories {
		mod := modFactory(core)
		if mod == nil {
			panic("proposal module factory returned nil")
		}

		core.proposalModules[i] = dao_interfaces.ActivableProposalModule{
			Enabled: true,
			Module:  mod,
		}
	}

	// this registry is specific to gno since we can't do dynamic calls
	core.registry.Register(NewUpdateVotingModuleMessageHandler(core))
	core.registry.Register(NewUpdateProposalModulesMessageHandler(core))
	for _, handlerFactory := range messageHandlersFactories {
		handler := handlerFactory(core)
		if handler == nil {
			panic("message handler factory returned nil")
		}

		core.registry.Register(handler)
	}

	return core
}

// mutations

func (d *daoCore) UpdateVotingModule(newVotingModule dao_interfaces.IVotingModule) {
	if std.CurrentRealm().Addr() != d.realm.Addr() { // not sure this check necessary since the ownership system should protect against mutation from other realms
		panic(ErrUnauthorized)
	}

	d.votingModule = newVotingModule
}

func (d *daoCore) UpdateProposalModules(toAdd []dao_interfaces.IProposalModule, toDisable []int) {
	if std.CurrentRealm().Addr() != d.realm.Addr() { // not sure this check necessary since the ownership system should protect against mutation from other realms
		panic(ErrUnauthorized)
	}

	for _, module := range toAdd {
		d.addProposalModule(module)
	}

	for _, moduleIndex := range toDisable {
		module := GetProposalModule(d, moduleIndex)

		if !module.Enabled {
			panic(ErrModuleAlreadyDisabled)
		}

		module.Enabled = false

		d.activeProposalModuleCount--
		if d.activeProposalModuleCount == 0 {
			panic("no active proposal modules") // this -> `panic(ErrNoActiveProposalModules)` triggers `panic: reflect: reflect.Value.SetString using value obtained using unexported field`
		}
	}
}

// queries

func (d *daoCore) ProposalModules() []dao_interfaces.ActivableProposalModule {
	return d.proposalModules
}

func (d *daoCore) VotingModule() dao_interfaces.IVotingModule {
	return d.votingModule
}

func (d *daoCore) VotingPowerAtHeight(address std.Address, height int64) uint64 {
	return d.VotingModule().VotingPowerAtHeight(address, height, []string{})
}

func (d *daoCore) ActiveProposalModuleCount() int {
	return d.activeProposalModuleCount
}

func (d *daoCore) Render(path string) string {
	sb := strings.Builder{}
	sb.WriteString("# DAO Core\n")
	votingInfo := d.votingModule.Info()
	sb.WriteString("## Voting Module: ")
	sb.WriteString(votingInfo.String())
	sb.WriteRune('\n')
	sb.WriteString(d.votingModule.Render(""))
	sb.WriteString("## Supported Messages:\n")
	sb.WriteString(d.registry.Render())

	sb.WriteString("## Proposal Modules:\n")
	for i, propMod := range d.proposalModules {
		if !propMod.Enabled {
			continue
		}

		info := propMod.Module.Info()
		sb.WriteString("### #")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(": ")
		sb.WriteString(info.String())
		sb.WriteRune('\n')
		sb.WriteString(propMod.Module.Render(""))
	}

	sb.WriteRune('\n')
	return sb.String()
}

func (d *daoCore) Registry() *dao_interfaces.MessagesRegistry {
	return d.registry
}

// TODO: move this helper in dao interfaces

func GetProposalModule(core dao_interfaces.IDAOCore, moduleIndex int) *dao_interfaces.ActivableProposalModule {
	if moduleIndex < 0 {
		panic("module index must be >= 0")
	}

	mods := core.ProposalModules()
	if moduleIndex >= len(mods) {
		panic("invalid module index")
	}

	return &mods[moduleIndex]
}

// internal

func (d *daoCore) executeMsgs(msgs []dao_interfaces.ExecutableMessage) {
	for _, msg := range msgs {
		d.registry.Execute(msg)
	}
}

func (d *daoCore) addProposalModule(proposalMod dao_interfaces.IProposalModule) {
	for _, mod := range d.proposalModules {
		if mod.Module != proposalMod {
			continue
		}

		if mod.Enabled {
			panic(ErrModuleAlreadyAdded)
		}

		mod.Enabled = true
		d.activeProposalModuleCount++
		return
	}

	d.proposalModules = append(d.proposalModules, dao_interfaces.ActivableProposalModule{
		Enabled: true,
		Module:  proposalMod,
	})

	d.activeProposalModuleCount++
}
