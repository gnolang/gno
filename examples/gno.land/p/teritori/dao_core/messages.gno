package core

import (
	"gno.land/p/demo/json"
	dao_interfaces "gno.land/p/teritori/dao_interfaces"
)

// UpdateProposalModules

type UpdateProposalModulesExecutableMessage struct {
	ToAdd     []dao_interfaces.IProposalModule
	ToDisable []int
}

var _ dao_interfaces.ExecutableMessage = &UpdateProposalModulesExecutableMessage{}

func (msg UpdateProposalModulesExecutableMessage) Type() string {
	return "gno.land/p/teritori/dao_core.UpdateProposalModules"
}

func (msg *UpdateProposalModulesExecutableMessage) String() string {
	panic(ErrNotImplemented)
}

func (msg *UpdateProposalModulesExecutableMessage) ToJSON() *json.Node {
	panic(ErrNotImplemented)
}

func (msg *UpdateProposalModulesExecutableMessage) FromJSON(ast *json.Node) {
	panic(ErrNotImplemented)
}

type UpdateProposalModulesMessageHandler struct {
	dao dao_interfaces.IDAOCore
}

var _ dao_interfaces.MessageHandler = &UpdateProposalModulesMessageHandler{}

func NewUpdateProposalModulesMessageHandler(dao dao_interfaces.IDAOCore) *UpdateProposalModulesMessageHandler {
	return &UpdateProposalModulesMessageHandler{dao: dao}
}

func (handler UpdateProposalModulesMessageHandler) Type() string {
	return UpdateProposalModulesExecutableMessage{}.Type()
}

func (handler *UpdateProposalModulesMessageHandler) Execute(message dao_interfaces.ExecutableMessage) {
	msg := message.(*UpdateProposalModulesExecutableMessage)
	handler.dao.UpdateProposalModules(msg.ToAdd, msg.ToDisable)
}

func (handler *UpdateProposalModulesMessageHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &UpdateProposalModulesExecutableMessage{}
}

// UpdateVotingModule

type UpdateVotingModuleExecutableMessage struct {
	Module dao_interfaces.IVotingModule
}

var _ dao_interfaces.ExecutableMessage = &UpdateVotingModuleExecutableMessage{}

func (msg UpdateVotingModuleExecutableMessage) Type() string {
	return "gno.land/p/teritori/dao_core.UpdateVotingModule"
}

func (msg *UpdateVotingModuleExecutableMessage) String() string {
	panic(ErrNotImplemented)
}

func (msg *UpdateVotingModuleExecutableMessage) ToJSON() *json.Node {
	panic(ErrNotImplemented)
}

func (msg *UpdateVotingModuleExecutableMessage) FromJSON(ast *json.Node) {
	panic(ErrNotImplemented)
}

type UpdateVotingModuleMessageHandler struct {
	dao dao_interfaces.IDAOCore
}

var _ dao_interfaces.MessageHandler = &UpdateVotingModuleMessageHandler{}

func NewUpdateVotingModuleMessageHandler(dao dao_interfaces.IDAOCore) *UpdateVotingModuleMessageHandler {
	return &UpdateVotingModuleMessageHandler{dao: dao}
}

func (handler UpdateVotingModuleMessageHandler) Type() string {
	return UpdateVotingModuleExecutableMessage{}.Type()
}

func (handler *UpdateVotingModuleMessageHandler) Execute(message dao_interfaces.ExecutableMessage) {
	msg := message.(*UpdateVotingModuleExecutableMessage)
	handler.dao.UpdateVotingModule(msg.Module)
}

func (handler *UpdateVotingModuleMessageHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &UpdateVotingModuleExecutableMessage{}
}
