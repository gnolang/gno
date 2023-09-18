package dao_interfaces

import (
	"encoding/base64"
	"encoding/binary"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ujson"
)

type ExecutableMessage interface {
	ujson.JSONAble
	ujson.FromJSONAble

	String() string
	Type() string
}

type MessageHandler interface {
	Execute(message ExecutableMessage)
	MessageFromJSON(ast *ujson.JSONASTNode) ExecutableMessage
	Type() string
}

type MessageHandlerFactory func(core IDAOCore) MessageHandler
