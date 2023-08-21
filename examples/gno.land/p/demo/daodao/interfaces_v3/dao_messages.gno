package dao_interfaces

import (
	"encoding/base64"
	"encoding/binary"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/jsonutil_v3"
)

type MessageHandler interface {
	Execute(message ExecutableMessage)
	FromBinary(b []byte) ExecutableMessage
	FromJSON(ast *jsonutil.JSONASTNode) ExecutableMessage
	Type() string
}

type MessagesRegistry struct {
	handlers *avl.Tree
}

func NewMessagesRegistry() *MessagesRegistry {
	return &MessagesRegistry{handlers: avl.NewTree()}
}

func (r *MessagesRegistry) Register(handler MessageHandler) {
	r.handlers.Set(handler.Type(), handler)
}

func (r *MessagesRegistry) FromBinary(b []byte) ExecutableMessage {
	if len(b) < 2 {
		panic("invalid ExecutableMessage: invalid length")
	}
	l := binary.BigEndian.Uint16(b[:2])
	if len(b) < int(l+2) {
		panic("invalid ExecutableMessage: invalid length")
	}
	t := string(b[2 : l+2])

	h, ok := r.handlers.Get(t)
	if !ok {
		panic("invalid ExecutableMessage: invalid message type")
	}
	return h.(MessageHandler).FromBinary(b)
}

func (r *MessagesRegistry) FromJSON(s string) []ExecutableMessage {
	ast := jsonutil.TokenizeAndParse(s)
	if ast.Kind != jsonutil.JSONKindArray {
		panic("msgs is not an array")
	}
	msgs := make([]ExecutableMessage, 0, len(ast.ArrayChildren))
	for _, child := range ast.ArrayChildren {
		if child.Kind != jsonutil.JSONKindObject {
			panic("invalid ExecutableMessage: message container is not an object")
		}
		var messageType string
		var payload *jsonutil.JSONASTNode
		jsonutil.ParseObjectAST(child, []*jsonutil.ParseKV{
			{Key: "type", Value: &messageType},
			{Key: "payload", Value: &payload},
		})
		h, ok := r.handlers.Get(messageType)
		if !ok {
			panic("invalid ExecutableMessage: invalid message type")
		}
		msgs = append(msgs, h.(MessageHandler).FromJSON(payload))
	}
	return msgs
}

func (r *MessagesRegistry) FromBase64String(s string) ExecutableMessage {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		panic("invalid ExecutableMessage: invalid base64 string")
	}
	return r.FromBinary(b)
}

func (r *MessagesRegistry) Execute(msg ExecutableMessage) {
	h, ok := r.handlers.Get(msg.Type())
	if !ok {
		panic("invalid ExecutableMessage: invalid message type")
	}
	return h.(MessageHandler).Execute(msg)
}

func (r *MessagesRegistry) ExecuteMessages(msgs []ExecutableMessage) {
	for _, msg := range msgs {
		r.Execute(msg)
	}
}
