package groups

import (
	"encoding/binary"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/daodao/interfaces"
	"gno.land/p/demo/ujson"
)

type ExecutableMessageAddMember struct {
	dao_interfaces.ExecutableMessage

	GroupID  GroupID
	Address  string
	Weight   int
	Metadata string
}

func (msg *ExecutableMessageAddMember) Type() string {
	return "gno.land/r/demo/groups.AddMember"
}

func (msg *ExecutableMessageAddMember) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	s := "GroupID: " + msg.GroupID.String() + "\n"
	s += "Address: " + msg.Address + "\n"
	s += "Weight: " + strconv.Itoa(msg.Weight) + "\n"
	s += "Metadata: " + msg.Metadata
	ss = append(ss, s)
	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageAddMember) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "groupId", Value: msg.GroupID},
		{Key: "address", Value: msg.Address},
		{Key: "weight", Value: msg.Weight},
		{Key: "metadata", Value: msg.Metadata},
	})
}

type AddMemberHandler struct {
	dao_interfaces.MessageHandler
}

func NewAddMemberHandler() *AddMemberHandler {
	return &AddMemberHandler{}
}

func (h *AddMemberHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageAddMember)
	AddMember(msg.GroupID, msg.Address, msg.Weight, msg.Metadata)
}

func (h *AddMemberHandler) Type() string {
	return ExecutableMessageAddMember{}.Type()
}

func (h *AddMemberHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	msg := &ExecutableMessageAddMember{}
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "groupId", Value: &msg.GroupID},
		{Key: "address", Value: &msg.Address},
		{Key: "weight", Value: &msg.Weight},
		{Key: "metadata", Value: &msg.Metadata},
	})
	return msg
}

type ExecutableMessageDeleteMember struct {
	dao_interfaces.ExecutableMessage

	GroupID  GroupID
	MemberID MemberID
}

func (msg *ExecutableMessageDeleteMember) Type() string {
	return "gno.land/r/demo/groups.DeleteMember"
}

func (msg *ExecutableMessageDeleteMember) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	s := "GroupID: " + msg.GroupID.String()
	s += "MemberID: " + msg.MemberID.String()
	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageDeleteMember) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "groupId", Value: msg.GroupID},
		{Key: "memberId", Value: msg.MemberID},
	})
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

func (h *DeleteMemberHandler) Type() string {
	return ExecutableMessageDeleteMember{}.Type()
}

func (h *DeleteMemberHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	msg := &ExecutableMessageDeleteMember{}
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "groupId", Value: &msg.GroupID},
		{Key: "memberId", Value: &msg.MemberID},
	})
	return msg
}
