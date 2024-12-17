package dao_voting_group

import (
	"gno.land/p/teritori/dao_interfaces"
	"gno.land/p/demo/json"
)

const updateMembersType = "gno.land/p/teritori/dao_voting_group.UpdateMembers"

type UpdateMembersExecutableMessage []Member

var _ dao_interfaces.ExecutableMessage = (*UpdateMembersExecutableMessage)(nil)

func (m *UpdateMembersExecutableMessage) FromJSON(ast *json.Node) {
	changes := ast.MustArray()
	*m = make([]Member, len(changes))
	for i, change := range changes {
		(*m)[i].FromJSON(change)
	}
}

func (m *UpdateMembersExecutableMessage) ToJSON() *json.Node {
	changes := make([]*json.Node, len(*m))
	for i, change := range *m {
		changes[i] = change.ToJSON()
	}

	return json.ArrayNode("", changes)
}

func (m *UpdateMembersExecutableMessage) String() string {
	return m.ToJSON().String()
}

func (m *UpdateMembersExecutableMessage) Type() string {
	return updateMembersType
}

type updateMembersHandler struct {
	vg *VotingGroup
}

var _ dao_interfaces.MessageHandler = (*updateMembersHandler)(nil)

func (h *updateMembersHandler) Type() string {
	return updateMembersType
}

func (h *updateMembersHandler) Execute(msg dao_interfaces.ExecutableMessage) {
	m, ok := msg.(*UpdateMembersExecutableMessage)
	if !ok {
		panic("unexpected message type")
	}

	for _, change := range *m {
		h.vg.SetMemberPower(change.Address, change.Power)
	}
}

func (h *updateMembersHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &UpdateMembersExecutableMessage{}
}
