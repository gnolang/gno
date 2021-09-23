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

// NewMsgAddPackage - execute a Gno statement.
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
// MsgExec

// MsgExec - execute a Gno statement.
type MsgExec struct {
	Caller crypto.Address `json:"caller" yaml:"caller"`
	Stmt   string         `json:"stmt" yaml:"stmt"`
}

var _ std.Msg = MsgExec{}

// NewMsgExec - execute a Gno statement.
func NewMsgExec(caller crypto.Address, stmt string) MsgExec {
	return MsgExec{
		Caller: caller,
		Stmt:   stmt,
	}
}

// Route Implements Msg.
func (msg MsgExec) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgExec) Type() string { return "exec" }

// ValidateBasic Implements Msg.
func (msg MsgExec) ValidateBasic() error {
	if msg.Caller.IsZero() {
		return std.ErrInvalidAddress("missing caller address")
	}
	if msg.Stmt == "" { // XXX
		return ErrInvalidStmt("missing statement to execute")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgExec) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgExec) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Caller}
}
