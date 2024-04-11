package groups

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/json"
	"gno.land/p/demo/teritori/dao_interfaces"
	"gno.land/p/demo/teritori/jsonutil"
)

type ExecutableMessageAddMember struct {
	dao_interfaces.ExecutableMessage

	GroupID  GroupID
	Address  std.Address
	Weight   int
	Metadata string
}

var _ dao_interfaces.ExecutableMessage = (*ExecutableMessageAddMember)(nil)

func (msg ExecutableMessageAddMember) Type() string {
	return "gno.land/r/demo/teritori/groups.AddMember"
}

func (msg *ExecutableMessageAddMember) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	s := "GroupID: " + msg.GroupID.String() + "\n"
	s += "Address: " + msg.Address.String() + "\n"
	s += "Weight: " + strconv.Itoa(msg.Weight) + "\n"
	s += "Metadata: " + msg.Metadata
	ss = append(ss, s)
	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageAddMember) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"groupId":  msg.GroupID.ToJSON(),
		"address":  jsonutil.AddressNode(msg.Address),
		"weight":   jsonutil.IntNode(msg.Weight),
		"metadata": json.StringNode("", msg.Metadata),
	})
}

func (msg *ExecutableMessageAddMember) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	msg.GroupID.FromJSON(obj["groupId"])
	msg.Address = jsonutil.MustAddress(obj["address"])
	msg.Weight = jsonutil.MustInt(obj["weight"])
	msg.Metadata = obj["metadata"].MustString()
}

type AddMemberHandler struct {
	dao_interfaces.MessageHandler
}

func NewAddMemberHandler() *AddMemberHandler {
	return &AddMemberHandler{}
}

func (h *AddMemberHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageAddMember)
	AddMember(msg.GroupID, msg.Address.String(), msg.Weight, msg.Metadata)
}

func (h AddMemberHandler) Type() string {
	return ExecutableMessageAddMember{}.Type()
}

func (h *AddMemberHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &ExecutableMessageAddMember{}
}

type ExecutableMessageDeleteMember struct {
	dao_interfaces.ExecutableMessage

	GroupID  GroupID
	MemberID MemberID
}

var _ dao_interfaces.ExecutableMessage = (*ExecutableMessageDeleteMember)(nil)

func (msg ExecutableMessageDeleteMember) Type() string {
	return "gno.land/r/demo/teritori/groups.DeleteMember"
}

func (msg *ExecutableMessageDeleteMember) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	s := "GroupID: " + msg.GroupID.String()
	s += "MemberID: " + msg.MemberID.String()
	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageDeleteMember) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"groupId":  msg.GroupID.ToJSON(),
		"memberId": msg.MemberID.ToJSON(),
	})
}

func (msg *ExecutableMessageDeleteMember) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	msg.GroupID.FromJSON(obj["groupId"])
	msg.MemberID.FromJSON(obj["memberId"])
}

type DeleteMemberHandler struct {
	dao_interfaces.MessageHandler
}

func NewDeleteMemberHandler() *DeleteMemberHandler {
	return &DeleteMemberHandler{}
}

func (h *DeleteMemberHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageDeleteMember)
	DeleteMember(msg.GroupID, msg.MemberID)
}

func (h DeleteMemberHandler) Type() string {
	return ExecutableMessageDeleteMember{}.Type()
}

func (h *DeleteMemberHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &ExecutableMessageDeleteMember{}
}
