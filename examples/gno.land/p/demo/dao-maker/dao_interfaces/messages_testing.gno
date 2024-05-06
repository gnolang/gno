package dao_interfaces

import (
	"gno.land/p/demo/json"
)

type CopyMessage struct {
	Value string
}

func (m CopyMessage) Type() string {
	return "CopyMessage"
}

func (m *CopyMessage) String() string {
	return m.Value
}

func (m *CopyMessage) FromJSON(ast *json.Node) {
	m.Value = ast.MustString()
}

func (m *CopyMessage) ToJSON() *json.Node {
	return json.StringNode("", m.Value)
}

type CopyMessageHandler struct {
	ptr *string
}

func NewCopyMessageHandler(ptr *string) *CopyMessageHandler {
	if ptr == nil {
		panic("ptr cannot be nil")
	}

	return &CopyMessageHandler{ptr}
}

func (h *CopyMessageHandler) Execute(imsg ExecutableMessage) {
	msg, ok := imsg.(*CopyMessage)
	if !ok {
		panic("Wrong message type")
	}

	*h.ptr = msg.Value
}

func (h CopyMessageHandler) Type() string {
	return "CopyMessage"
}

func (h *CopyMessageHandler) Instantiate() ExecutableMessage {
	return &CopyMessage{}
}
