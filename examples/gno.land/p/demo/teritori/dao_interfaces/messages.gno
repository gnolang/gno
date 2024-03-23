package dao_interfaces

import (
	"gno.land/p/demo/teritori/ujson"
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
