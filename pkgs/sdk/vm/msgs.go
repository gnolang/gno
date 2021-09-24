package vm

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"
)

type NamedFile struct {
	Name string
	Body string
}

//----------------------------------------
// MsgAddPackage

// MsgAddPackage - create and initialize new package
type MsgAddPackage struct {
	Creator crypto.Address `json:"creator" yaml:"creator"`
	PkgPath string         `json:"pkg_path" yaml:"pkg_path"`
	Files   []NamedFile    `json:"files" yaml:"files"`
}

var _ std.Msg = MsgAddPackage{}

// NewMsgAddPackage - upload a package with files.
func NewMsgAddPackage(creator crypto.Address, pkgPath string, files []NamedFile) MsgAddPackage {
	return MsgAddPackage{
		Creator: creator,
		PkgPath: pkgPath,
		Files:   files,
	}
}

// Route Implements Msg.
func (msg MsgAddPackage) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgAddPackage) Type() string { return "add_package" }

// ValidateBasic Implements Msg.
func (msg MsgAddPackage) ValidateBasic() error {
	if msg.Creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	if msg.PkgPath == "" { // XXX
		return ErrInvalidPkgPath("missing package path")
	}
	// XXX validate files.
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgAddPackage) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgAddPackage) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Creator}
}

//----------------------------------------
// MsgEval

// MsgEval - evaluate a Gno expression.
type MsgEval struct {
	Caller  crypto.Address `json:"caller" yaml:"caller"`
	PkgPath string         `json:"pkg_path" yaml:"pkg_path"`
	Expr    string         `json:"expr" yaml:"expr"`
}

var _ std.Msg = MsgEval{}

// NewMsgEval - evaluate a Gno expression.
func NewMsgEval(caller crypto.Address, pkgPath, expr string) MsgEval {
	return MsgEval{
		Caller:  caller,
		PkgPath: pkgPath,
		Expr:    expr,
	}
}

// Route Implements Msg.
func (msg MsgEval) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgEval) Type() string { return "exec" }

// ValidateBasic Implements Msg.
func (msg MsgEval) ValidateBasic() error {
	if msg.Caller.IsZero() {
		return std.ErrInvalidAddress("missing caller address")
	}
	if msg.PkgPath == "" { // XXX
		return ErrInvalidPkgPath("missing package path")
	}
	if msg.Expr == "" { // XXX
		return ErrInvalidExpr("missing expression to evaluate")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgEval) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgEval) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Caller}
}
