package tori

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/json"
	"gno.land/p/demo/users"
	"gno.land/p/teritori/dao_interfaces"
	"gno.land/p/teritori/jsonutil"
)

// TODO: move this file in a generic package to administrate grc20s via daos

type ExecutableMessageMintTori struct {
	dao_interfaces.ExecutableMessage

	Recipient std.Address
	Amount    uint64
}

var _ dao_interfaces.ExecutableMessage = &ExecutableMessageMintTori{}

func (msg ExecutableMessageMintTori) Type() string {
	return "gno.land/r/teritori/tori.MintTori"
}

func (msg *ExecutableMessageMintTori) String() string {
	sb := strings.Builder{}
	sb.WriteString(msg.Type())
	sb.WriteString("\n---\n")
	sb.WriteString("Recipient: ")
	sb.WriteString(string(msg.Recipient))
	sb.WriteRune('\n')
	sb.WriteString("Amount: ")
	sb.WriteString(strconv.FormatUint(msg.Amount, 10))
	return sb.String()
}

func (msg *ExecutableMessageMintTori) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	msg.Recipient = jsonutil.MustAddress(obj["recipient"])
	msg.Amount = jsonutil.MustUint64(obj["amount"])
}

func (msg *ExecutableMessageMintTori) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"recipient": jsonutil.AddressNode(msg.Recipient),
		"amount":    jsonutil.Uint64Node(msg.Amount),
	})
}

type MintToriHandler struct {
	dao_interfaces.MessageHandler
}

var _ dao_interfaces.MessageHandler = &MintToriHandler{}

func NewMintToriHandler() *MintToriHandler {
	return &MintToriHandler{}
}

func (h *MintToriHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageMintTori)
	Mint(users.AddressOrName(msg.Recipient), msg.Amount)
}

func (h MintToriHandler) Type() string {
	return ExecutableMessageMintTori{}.Type()
}

func (h *MintToriHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &ExecutableMessageMintTori{}
}

type ExecutableMessageBurnTori struct {
	dao_interfaces.ExecutableMessage

	Target users.AddressOrName
	Amount uint64
}

var _ dao_interfaces.ExecutableMessage = &ExecutableMessageBurnTori{}

func (msg ExecutableMessageBurnTori) Type() string {
	return "gno.land/r/teritori/tori.BurnTori"
}

func (msg *ExecutableMessageBurnTori) String() string {
	sb := strings.Builder{}
	sb.WriteString(msg.Type())
	sb.WriteString("\n---\n")
	sb.WriteString("Target: ")
	sb.WriteString(string(msg.Target))
	sb.WriteRune('\n')
	sb.WriteString("Amount: ")
	sb.WriteString(strconv.FormatUint(msg.Amount, 10))
	return sb.String()
}

func (msg *ExecutableMessageBurnTori) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	msg.Target = jsonutil.MustAddressOrName(obj["target"])
	msg.Amount = jsonutil.MustUint64(obj["amount"])
}

func (msg *ExecutableMessageBurnTori) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"target": jsonutil.AddressOrNameNode(msg.Target),
		"amount": jsonutil.Uint64Node(msg.Amount),
	})
}

type BurnToriHandler struct {
	dao_interfaces.MessageHandler
}

var _ dao_interfaces.MessageHandler = &BurnToriHandler{}

func NewBurnToriHandler() *BurnToriHandler {
	return &BurnToriHandler{}
}

func (h *BurnToriHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageBurnTori)
	Burn(msg.Target, msg.Amount)
}

func (h BurnToriHandler) Type() string {
	return ExecutableMessageBurnTori{}.Type()
}

func (h *BurnToriHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &ExecutableMessageBurnTori{}
}

type ExecutableMessageChangeAdmin struct {
	dao_interfaces.ExecutableMessage

	NewAdmin std.Address
}

var _ dao_interfaces.ExecutableMessage = &ExecutableMessageChangeAdmin{}

func (msg ExecutableMessageChangeAdmin) Type() string {
	return "gno.land/r/teritori/tori.ChangeAdmin"
}

func (msg *ExecutableMessageChangeAdmin) String() string {
	var ss []string
	ss = append(ss, msg.Type())
	s := "New admin: " + string(msg.NewAdmin)
	ss = append(ss, s)
	return strings.Join(ss, "\n---\n")
}

func (msg *ExecutableMessageChangeAdmin) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	msg.NewAdmin = jsonutil.MustAddress(obj["newAdmin"])
}

func (msg *ExecutableMessageChangeAdmin) ToJSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"newAdmin": jsonutil.AddressNode(msg.NewAdmin),
	})
}

type ChangeAdminHandler struct {
	dao_interfaces.MessageHandler
}

var _ dao_interfaces.MessageHandler = &ChangeAdminHandler{}

func NewChangeAdminHandler() *ChangeAdminHandler {
	return &ChangeAdminHandler{}
}

func (h *ChangeAdminHandler) Execute(imsg dao_interfaces.ExecutableMessage) {
	msg := imsg.(*ExecutableMessageChangeAdmin)
	owner.TransferOwnership(msg.NewAdmin)
}

func (h ChangeAdminHandler) Type() string {
	return ExecutableMessageChangeAdmin{}.Type()
}

func (h *ChangeAdminHandler) Instantiate() dao_interfaces.ExecutableMessage {
	return &ExecutableMessageChangeAdmin{}
}
