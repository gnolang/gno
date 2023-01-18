package vm

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/std"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/sdk/vm",
	"vm",
	amino.GetCallersDirname(),
).WithDependencies(
	std.Package,
).WithTypes(
	MsgCall{}, "m_call",
	MsgAddPackage{}, "m_addpkg", // TODO rename both to MsgAddPkg?
	MsgEval{}, "m_eval",

	// errors
	InvalidPkgPathError{}, "InvalidPkgPathError",
	InvalidStmtError{}, "InvalidStmtError",
	InvalidExprError{}, "InvalidExprError",
))
