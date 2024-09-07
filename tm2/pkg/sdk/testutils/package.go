package testutils

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/sdk/testutils",
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
