package vm

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm",
	"vm",
	amino.GetCallersDirname(),
).WithDependencies(
	std.Package,
	params.Package,
).WithTypes(
	MsgCall{}, "m_call",
	MsgRun{}, "m_run",
	MsgAddPackage{}, "m_addpkg", // TODO rename both to MsgAddPkg?
	MsgEnablePackage{}, "m_enable_pkg",
	MsgDisablePackage{}, "m_disable_pkg",
	MsgRejectPackage{}, "m_reject_pkg",

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
	ObjectNotFoundError{}, "ObjectNotFoundError",
	ExportSizeExceededError{}, "ExportSizeExceededError",
	ExportDepthExceededError{}, "ExportDepthExceededError",
	UnobservedSendError{}, "UnobservedSendError",
	UnspendableSendError{}, "UnspendableSendError",
	GenesisState{}, "GenesisState",
	Params{}, "Params",
))
