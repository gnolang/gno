package vm

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/sdk/vm",
	"vm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	MsgEval{}, "m_eval",
	MsgAddPackage{}, "m_addpkg",
))
