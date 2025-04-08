package auth

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/sdk/auth",
	"auth",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	Event{}, "Event",
	EventAttribute{}, "EventAttribute",
))
