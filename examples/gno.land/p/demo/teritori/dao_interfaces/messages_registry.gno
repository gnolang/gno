package dao_interfaces

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/teritori/ujson"
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

func (r *MessagesRegistry) MessagesFromJSON(messagesJSON string) []ExecutableMessage {
	slice := ujson.ParseSlice(messagesJSON)
	msgs := make([]ExecutableMessage, 0, len(slice))
	for _, child := range slice {
		var messageType string
		var payload *ujson.JSONASTNode
		child.ParseObject([]*ujson.ParseKV{
			{Key: "type", Value: &messageType},
			{Key: "payload", Value: &payload},
		})
		h, ok := r.handlers.Get(messageType)
		if !ok {
			panic("invalid ExecutableMessage: invalid message type")
		}
		msgs = append(msgs, h.(MessageHandler).MessageFromJSON(payload))
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

type RegisterHandlerExecutableMessage struct {
	Handler MessageHandler
}

func (m RegisterHandlerExecutableMessage) Type() string {
	return "gno.land/p/demo/teritori/dao_interfaces.RegisterHandler"
}

func (m *RegisterHandlerExecutableMessage) FromJSON(ast *ujson.JSONASTNode) {
	panic("not implemented")
}

func (m *RegisterHandlerExecutableMessage) ToJSON() string {
	panic("not implemented")
}

func (m *RegisterHandlerExecutableMessage) String() string {
	return m.Handler.Type()
}

type RegisterHandlerExecutableMessageHandler struct {
	registry *MessagesRegistry
}

func NewRegisterHandlerExecutableMessageHandler(registry *MessagesRegistry) *RegisterHandlerExecutableMessageHandler {
	return &RegisterHandlerExecutableMessageHandler{registry: registry}
}

func (h RegisterHandlerExecutableMessageHandler) Type() string {
	return RegisterHandlerExecutableMessage{}.Type()
}

func (h *RegisterHandlerExecutableMessageHandler) MessageFromJSON(ast *ujson.JSONASTNode) ExecutableMessage {
	panic("not implemented")
}

func (h *RegisterHandlerExecutableMessageHandler) Execute(msg ExecutableMessage) {
	h.registry.Register(msg.(*RegisterHandlerExecutableMessage).Handler)
}

type RemoveHandlerExecutableMessage struct {
	HandlerType string
}

func (m RemoveHandlerExecutableMessage) Type() string {
	return "gno.land/p/demo/teritori/dao_interfaces.RemoveHandler"
}

func (m *RemoveHandlerExecutableMessage) FromJSON(ast *ujson.JSONASTNode) {
	ast.ParseAny(&m.HandlerType)
}

func (m *RemoveHandlerExecutableMessage) ToJSON() string {
	return ujson.FormatAny(m.HandlerType)
}

func (m *RemoveHandlerExecutableMessage) String() string {
	return m.HandlerType
}

type RemoveHandlerExecutableMessageHandler struct {
	registry *MessagesRegistry
}

func NewRemoveHandlerExecutableMessageHandler(registry *MessagesRegistry) *RemoveHandlerExecutableMessageHandler {
	return &RemoveHandlerExecutableMessageHandler{registry: registry}
}

func (h RemoveHandlerExecutableMessageHandler) Type() string {
	return RemoveHandlerExecutableMessage{}.Type()
}

func (h *RemoveHandlerExecutableMessageHandler) MessageFromJSON(ast *ujson.JSONASTNode) ExecutableMessage {
	msg := &RemoveHandlerExecutableMessage{}
	ast.ParseAny(msg)
	return msg
}

func (h *RemoveHandlerExecutableMessageHandler) Execute(msg ExecutableMessage) {
	h.registry.Remove(msg.(*RemoveHandlerExecutableMessage).HandlerType)
}
