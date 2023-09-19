package core

import (
	dao_interfaces "gno.land/p/demo/teritori/dao_interfaces"
	"gno.land/p/demo/teritori/ujson"
)

// UpdateProposalModules

type UpdateProposalModulesExecutableMessage struct {
	ToAdd     []dao_interfaces.IProposalModule
	ToDisable []int
}

func (msg *UpdateProposalModulesExecutableMessage) Type() string {
	return "gno.land/p/demo/teritori/dao_core.UpdateProposalModules"
}

func (msg *UpdateProposalModulesExecutableMessage) String() string {
	panic(ErrNotImplemented)
}

func (msg *UpdateProposalModulesExecutableMessage) ToJSON() string {
	panic(ErrNotImplemented)
}

func (msg *UpdateProposalModulesExecutableMessage) FromJSON(ast *ujson.JSONASTNode) {
	panic(ErrNotImplemented)
}

type UpdateProposalModulesMessageHandler struct {
	dao dao_interfaces.IDAOCore
}

func NewUpdateProposalModulesMessageHandler(dao dao_interfaces.IDAOCore) *UpdateProposalModulesMessageHandler {
	return &UpdateProposalModulesMessageHandler{dao: dao}
}

func (handler *UpdateProposalModulesMessageHandler) Type() string {
	return UpdateProposalModulesExecutableMessage{}.Type()
}

func (handler *UpdateProposalModulesMessageHandler) Execute(message dao_interfaces.ExecutableMessage) {
	msg := message.(*UpdateProposalModulesExecutableMessage)
	handler.dao.UpdateProposalModules(msg.ToAdd, msg.ToDisable)
}

func (handler *UpdateProposalModulesMessageHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	panic(ErrNotSupported)
}

// UpdateVotingModule

type UpdateVotingModuleExecutableMessage struct {
	Module dao_interfaces.IVotingModule
}

func (msg *UpdateVotingModuleExecutableMessage) Type() string {
	return "gno.land/p/demo/teritori/dao_core.UpdateVotingModule"
}

func (msg *UpdateVotingModuleExecutableMessage) String() string {
	panic(ErrNotImplemented)
}

func (msg *UpdateVotingModuleExecutableMessage) ToJSON() string {
	panic(ErrNotImplemented)
}

func (msg *UpdateVotingModuleExecutableMessage) FromJSON(ast *ujson.JSONASTNode) {
	panic(ErrNotImplemented)
}

type UpdateVotingModuleMessageHandler struct {
	dao dao_interfaces.IDAOCore
}

func NewUpdateVotingModuleMessageHandler(dao dao_interfaces.IDAOCore) *UpdateVotingModuleMessageHandler {
	return &UpdateVotingModuleMessageHandler{dao: dao}
}

func (handler *UpdateVotingModuleMessageHandler) Type() string {
	return UpdateVotingModuleExecutableMessage{}.Type()
}

func (handler *UpdateVotingModuleMessageHandler) Execute(message dao_interfaces.ExecutableMessage) {
	msg := message.(*UpdateVotingModuleExecutableMessage)
	handler.dao.UpdateVotingModule(msg.Module)
}

func (handler *UpdateVotingModuleMessageHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	panic(ErrNotSupported)
}
