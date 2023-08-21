package groups

import (
	"encoding/binary"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/binutils"
	"gno.land/p/demo/daodao/interfaces_v3"
	"gno.land/p/demo/jsonutil_v3"
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

func (msg *ExecutableMessageAddMember) Binary() []byte {
	b := []byte{}
	b = append(b, binutils.EncodeLengthPrefixedStringUint16BE(msg.Type())...)
	b = binary.BigEndian.AppendUint64(b, uint64(msg.GroupID))
	b = append(b, binutils.EncodeLengthPrefixedStringUint16BE(msg.Address)...)
	b = binary.BigEndian.AppendUint32(b, uint32(msg.Weight))
	b = append(b, binutils.EncodeLengthPrefixedStringUint16BE(msg.Metadata)...)
	return b
}

func ExecutableMessageAddMemberFromBinary(b []byte) *ExecutableMessageAddMember {
	msg := &ExecutableMessageAddMember{}
	t, b := binutils.MustDecodeLengthPrefixedStringUint16BE(b)
	if t != msg.Type() {
		panic("invalid type")
	}
	msg.GroupID, b = GroupID(binary.BigEndian.Uint64(b)), b[8:]
	msg.Address, b = binutils.MustDecodeLengthPrefixedStringUint16BE(b)
	msg.Weight, b = int(binary.BigEndian.Uint32(b)), b[4:]
	msg.Metadata, b = binutils.MustDecodeLengthPrefixedStringUint16BE(b)
	return msg
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

func (h *AddMemberHandler) FromBinary(b []byte) dao_interfaces.ExecutableMessage {
	return ExecutableMessageAddMemberFromBinary(b)
}

func (h *AddMemberHandler) FromJSON(ast *jsonutil.JSONASTNode) dao_interfaces.ExecutableMessage {
	msg := &ExecutableMessageAddMember{}
	jsonutil.ParseObjectAST(ast, []*jsonutil.ParseKV{
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

func (msg *ExecutableMessageDeleteMember) Binary() []byte {
	b := []byte{}
	b = append(b, binutils.EncodeLengthPrefixedStringUint16BE(msg.Type())...)
	b = binary.BigEndian.AppendUint64(b, uint64(msg.GroupID))
	b = binary.BigEndian.AppendUint64(b, uint64(msg.MemberID))
	return b
}

func ExecutableMessageDeleteMemberFromBinary(b []byte) *ExecutableMessageDeleteMember {
	msg := &ExecutableMessageDeleteMember{}
	t, b := binutils.MustDecodeLengthPrefixedStringUint16BE(b)
	if t != msg.Type() {
		panic("invalid type")
	}
	msg.GroupID, b = GroupID(binary.BigEndian.Uint64(b)), b[8:]
	msg.MemberID, b = MemberID(binary.BigEndian.Uint64(b)), b[8:]
	return msg
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

func (h *DeleteMemberHandler) FromBinary(b []byte) dao_interfaces.ExecutableMessage {
	return ExecutableMessageDeleteMemberFromBinary(b)
}

func (h *DeleteMemberHandler) FromJSON(ast *jsonutil.JSONASTNode) dao_interfaces.ExecutableMessage {
	msg := &ExecutableMessageDeleteMember{}
	jsonutil.ParseObjectAST(ast, []*jsonutil.ParseKV{
		{Key: "groupId", Value: &msg.GroupID},
		{Key: "memberId", Value: &msg.MemberID},
	})
	return msg
}
