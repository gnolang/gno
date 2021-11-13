package vm

import (
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/std"
)

//----------------------------------------
// MsgAddPackage

// MsgAddPackage - create and initialize new package
type MsgAddPackage struct {
	Creator crypto.Address  `json:"creator" yaml:"creator"`
	Package std.MemPkgFiles `json:"package" yaml:"package"`
	Deposit std.Coins       `json:"deposit" yaml:"deposit"`
}

var _ std.Msg = MsgAddPackage{}

// NewMsgAddPackage - upload a package with files.
func NewMsgAddPackage(creator crypto.Address, pkgPath string, files []std.MemFile) MsgAddPackage {
	return MsgAddPackage{
		Creator: creator,
		Package: std.MemPkgFiles{
			Path:  pkgPath,
			Files: files,
		},
	}
}

// Implements Msg.
func (msg MsgAddPackage) Route() string { return RouterKey }

// Implements Msg.
func (msg MsgAddPackage) Type() string { return "add_package" }

// Implements Msg.
func (msg MsgAddPackage) ValidateBasic() error {
	if msg.Creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	if msg.Package.Path == "" { // XXX
		return ErrInvalidPkgPath("missing package path")
	}
	if !msg.Deposit.IsValid() {
		return std.ErrTxDecode("invalid deposit")
	}
	// XXX validate files.
	return nil
}

// Implements Msg.
func (msg MsgAddPackage) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// Implements Msg.
func (msg MsgAddPackage) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Creator}
}

// Implements ReceiveMsg.
func (msg MsgAddPackage) GetReceived() std.Coins {
	return msg.Deposit
}

//----------------------------------------
// MsgExec

// MsgExec - executes a Gno statement.
type MsgExec struct {
	Caller  crypto.Address `json:"caller" yaml:"caller"`
	PkgPath string         `json:"pkg_path" yaml:"pkg_path"`
	Stmt    string         `json:"stmt" yaml:"stmt"`
	Send    std.Coins      `json:"send" yaml:"send"`
}

var _ std.Msg = MsgExec{}

func NewMsgExec(caller crypto.Address, pkgPath, stmt string, send sdk.Coins) MsgExec {
	return MsgExec{
		Caller:  caller,
		PkgPath: pkgPath,
		Stmt:    stmt,
		Send:    send,
	}
}

// Implements Msg.
func (msg MsgExec) Route() string { return RouterKey }

// Implements Msg.
func (msg MsgExec) Type() string { return "exec" }

// Implements Msg.
func (msg MsgExec) ValidateBasic() error {
	if msg.Caller.IsZero() {
		return std.ErrInvalidAddress("missing caller address")
	}
	if msg.PkgPath == "" { // XXX
		return ErrInvalidPkgPath("missing package path")
	}
	if msg.Stmt == "" { // XXX
		return ErrInvalidExpr("missing expression to evaluate")
	}
	return nil
}

// Implements Msg.
func (msg MsgExec) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// Implements Msg.
func (msg MsgExec) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Caller}
}

// Implements ReceiveMsg.
func (msg MsgExec) GetReceived() std.Coins {
	return msg.Send
}
