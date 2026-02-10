package vm

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm",
	"vm",
	amino.GetCallersDirname(),
).WithDependencies(
	std.Package,
).WithTypes(
	MsgCall{}, "m_call",
	MsgRun{}, "m_run",
	MsgAddPackage{}, "m_addpkg", // TODO rename both to MsgAddPkg?

	// errors
	InvalidPkgPathError{}, "InvalidPkgPathError",
	NoRenderDeclError{}, "NoRenderDeclError",
	PkgExistError{}, "PkgExistError",
	InvalidStmtError{}, "InvalidStmtError",
	InvalidExprError{}, "InvalidExprError",
	TypeCheckError{}, "TypeCheckError",
	UnauthorizedUserError{}, "UnauthorizedUserError",
	InvalidPackageError{}, "InvalidPackageError",
	InvalidFileError{}, "InvalidFileError",
))
