package vm

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/sdk/vm",
	"vm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	MsgCall{}, "m_call",
	MsgAddPackage{}, "m_addpkg", // TODO rename both to MsgAddPkg?

	// errors
	InvalidPkgPathError{}, "InvalidPkgPathError",
	InvalidStmtError{}, "InvalidStmtError",
	InvalidExprError{}, "InvalidExprError",
))
