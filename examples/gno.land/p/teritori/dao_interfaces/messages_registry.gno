package dao_interfaces

import (
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"
)

type MessagesRegistry struct {
	handlers *avl.Tree
}

func NewMessagesRegistry() *MessagesRegistry {
	registry := &MessagesRegistry{handlers: avl.NewTree()}
	registry.Register(NewRegisterHandlerExecutableMessageHandler(registry))
	registry.Register(NewRemoveHandlerExecutableMessageHandler(registry))
	return registry
}

func (r *MessagesRegistry) Register(handler MessageHandler) {
	r.handlers.Set(handler.Type(), handler)
}

func (r *MessagesRegistry) Remove(t string) {
	r.handlers.Remove(t)
}

func (r *MessagesRegistry) MessagesFromJSON(slice []*json.Node) []ExecutableMessage {
	msgs := make([]ExecutableMessage, len(slice))
	for i, elem := range slice {
		messageType := json.Must(elem.GetKey("type")).MustString()
		payload := json.Must(elem.GetKey("payload"))
		h, ok := r.handlers.Get(messageType)
		if !ok {
			panic("invalid ExecutableMessage: invalid message type")
		}

		instance := h.(MessageHandler).Instantiate()
		instance.FromJSON(payload)
		msgs[i] = instance
	}

	return msgs
}

func (r *MessagesRegistry) Execute(msg ExecutableMessage) {
	h, ok := r.handlers.Get(msg.Type())
	if !ok {
		panic("invalid ExecutableMessage: invalid message type")
	}

	h.(MessageHandler).Execute(msg)
}

func (r *MessagesRegistry) ExecuteMessages(msgs []ExecutableMessage) {
	for _, msg := range msgs {
		r.Execute(msg)
	}
}

func (r *MessagesRegistry) Render() string {
	sb := strings.Builder{}
	r.handlers.Iterate("", "", func(key string, value interface{}) bool {
		sb.WriteString("- ")
		sb.WriteString(value.(MessageHandler).Type())
		sb.WriteRune('\n')
		return false
	})

	return sb.String()
}

type RegisterHandlerExecutableMessage struct {
	Handler MessageHandler
}

var _ ExecutableMessage = &RegisterHandlerExecutableMessage{}

func (m RegisterHandlerExecutableMessage) Type() string {
	return "gno.land/p/teritori/dao_interfaces.RegisterHandler"
}

func (m *RegisterHandlerExecutableMessage) FromJSON(ast *json.Node) {
	panic("not implemented")
}

func (m *RegisterHandlerExecutableMessage) ToJSON() *json.Node {
	panic("not implemented")
}

func (m *RegisterHandlerExecutableMessage) String() string {
	return m.Handler.Type()
}

type RegisterHandlerExecutableMessageHandler struct {
	registry *MessagesRegistry
}

var _ MessageHandler = &RegisterHandlerExecutableMessageHandler{}

func NewRegisterHandlerExecutableMessageHandler(registry *MessagesRegistry) *RegisterHandlerExecutableMessageHandler {
	return &RegisterHandlerExecutableMessageHandler{registry: registry}
}

func (h RegisterHandlerExecutableMessageHandler) Type() string {
	return RegisterHandlerExecutableMessage{}.Type()
}

func (h *RegisterHandlerExecutableMessageHandler) Instantiate() ExecutableMessage {
	return &RegisterHandlerExecutableMessage{}
}

func (h *RegisterHandlerExecutableMessageHandler) Execute(msg ExecutableMessage) {
	h.registry.Register(msg.(*RegisterHandlerExecutableMessage).Handler)
}

type RemoveHandlerExecutableMessage struct {
	HandlerType string
}

var _ ExecutableMessage = &RemoveHandlerExecutableMessage{}

func (m RemoveHandlerExecutableMessage) Type() string {
	return "gno.land/p/teritori/dao_interfaces.RemoveHandler"
}

func (m *RemoveHandlerExecutableMessage) FromJSON(ast *json.Node) {
	m.HandlerType = ast.MustString()
}

func (m *RemoveHandlerExecutableMessage) ToJSON() *json.Node {
	return json.StringNode("", m.HandlerType)
}

func (m *RemoveHandlerExecutableMessage) String() string {
	return m.HandlerType
}

type RemoveHandlerExecutableMessageHandler struct {
	registry *MessagesRegistry
}

var _ MessageHandler = &RemoveHandlerExecutableMessageHandler{}

func NewRemoveHandlerExecutableMessageHandler(registry *MessagesRegistry) *RemoveHandlerExecutableMessageHandler {
	return &RemoveHandlerExecutableMessageHandler{registry: registry}
}

func (h RemoveHandlerExecutableMessageHandler) Type() string {
	return RemoveHandlerExecutableMessage{}.Type()
}

func (h *RemoveHandlerExecutableMessageHandler) Instantiate() ExecutableMessage {
	return &RemoveHandlerExecutableMessage{}
}

func (h *RemoveHandlerExecutableMessageHandler) Execute(msg ExecutableMessage) {
	h.registry.Remove(msg.(*RemoveHandlerExecutableMessage).HandlerType)
}
