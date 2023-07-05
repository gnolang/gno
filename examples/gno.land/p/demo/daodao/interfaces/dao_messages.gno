package dao_interfaces

import (
	"encoding/base64"
	"encoding/binary"
	"strings"

	"gno.land/p/demo/cpavl"
)

type MessageHandler interface {
	Execute(message ExecutableMessage)
	FromBinary(b []byte) ExecutableMessage
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
