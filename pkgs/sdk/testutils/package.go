package testutils

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/sdk/testutils",
	"sdk.testutils",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	// ...
	&TestMsg{}, "TestMsg",

	// testmsgs.go
	MsgCounter{},
	MsgNoRoute{},
	MsgCounter2{},
))
