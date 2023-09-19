package tori

import (
	"encoding/binary"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/teritori/daodao/interfaces"
	"gno.land/p/demo/teritori/ujson"
	"gno.land/r/demo/users"
)

type ExecutableMessageMintTori struct {
	dao_interfaces.ExecutableMessage

	Address users.AddressOrName
	Amount  uint64
}

func (msg *ExecutableMessageMintTori) Type() string {
	return "gno.land/r/demo/tori.Mint"
}

func (msg *ExecutableMessageMintTori) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	s := "Address: " + string(msg.Address) + "\n"
	s += "Amount: " + strconv.FormatUint(msg.Amount, 10)
	ss = append(ss, s)
	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageMintTori) FromJSON(ast *ujson.JSONASTNode) {
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "address", Value: &msg.Address},
		{Key: "amount", Value: &msg.Amount},
	})
}

func (msg *ExecutableMessageMintTori) ToJSON() string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: "address", Value: msg.Address},
		{Key: "amount", Value: msg.Amount},
	})
}

type MintToriHandler struct {
	dao_interfaces.MessageHandler
}

func NewMintToriHandler() *MintToriHandler {
	return &MintToriHandler{}
}

func (h *MintToriHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageMintTori)
	Mint(msg.Address, msg.Amount)
}

func (h *MintToriHandler) Type() string {
	return ExecutableMessageMintTori{}.Type()
}

func (h *MintToriHandler) MessageFromJSON(ast *ujson.JSONASTNode) dao_interfaces.ExecutableMessage {
	msg := &ExecutableMessageMintTori{}
	msg.FromJSON(ast)
	return msg
}

// TODO: JSON

type ExecutableMessageBurnTori struct {
	dao_interfaces.ExecutableMessage

	Address users.AddressOrName
	Amount  uint64
}

func (msg *ExecutableMessageBurnTori) Type() string {
	return "BurnTori"
}

func (msg *ExecutableMessageBurnTori) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	s := "Address: " + string(msg.Address) + "\n"
	s += "Amount: " + strconv.FormatUint(msg.Amount, 10)
	ss = append(ss, s)
	return strings.Join(ss, "\n---\n")
}

type BurnToriHandler struct {
	dao_interfaces.MessageHandler
}

func NewBurnToriHandler() *BurnToriHandler {
	return &BurnToriHandler{}
}

func (h *BurnToriHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageBurnTori)
	Burn(msg.Address, msg.Amount)
}

func (h *BurnToriHandler) Type() string {
	return ExecutableMessageBurnTori{}.Type()
}
