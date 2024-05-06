package dao_interfaces

import (
	"gno.land/p/demo/json"
)

type ExecutableMessage interface {
	ToJSON() *json.Node
	FromJSON(ast *json.Node)

	String() string
	Type() string
}

type MessageHandler interface {
	Execute(message ExecutableMessage)
	Instantiate() ExecutableMessage
	Type() string
}

type MessageHandlerFactory func(core IDAOCore) MessageHandler
