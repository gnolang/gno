package vm

import (
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

//----------------------------------------
// MsgAddPackage

// MsgAddPackage - create and initialize new package
type MsgAddPackage struct {
	Creator crypto.Address  `json:"creator" yaml:"creator"`
	Package *std.MemPackage `json:"package" yaml:"package"`
	Deposit std.Coins       `json:"deposit" yaml:"deposit"`
}

var _ std.Msg = MsgAddPackage{}

// NewMsgAddPackage - upload a package with files.
func NewMsgAddPackage(creator crypto.Address, pkgPath string, files []*std.MemFile) MsgAddPackage {
	var pkgName string
	for _, file := range files {
		if strings.HasSuffix(file.Name, ".gno") {
			pkgName = string(gno.PackageNameFromFileBody(file.Name, file.Body))
			break
		}
	}
	return MsgAddPackage{
		Creator: creator,
		Package: &std.MemPackage{
			Name:  pkgName,
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
// MsgCall

// MsgCall - executes a Gno statement.
type MsgCall struct {
	Caller  crypto.Address `json:"caller" yaml:"caller"`
	Send    std.Coins      `json:"send" yaml:"send"`
	PkgPath string         `json:"pkg_path" yaml:"pkg_path"`
	Func    string         `json:"func" yaml:"func"`
	Args    []string       `json:"args" yaml:"args"`
}

var _ std.Msg = MsgCall{}

func NewMsgCall(caller crypto.Address, send sdk.Coins, pkgPath, fnc string, args []string) MsgCall {
	return MsgCall{
		Caller:  caller,
		Send:    send,
		PkgPath: pkgPath,
		Func:    fnc,
		Args:    args,
	}
}

// Implements Msg.
func (msg MsgCall) Route() string { return RouterKey }

// Implements Msg.
func (msg MsgCall) Type() string { return "exec" }

// Implements Msg.
func (msg MsgCall) ValidateBasic() error {
	if msg.Caller.IsZero() {
		return std.ErrInvalidAddress("missing caller address")
	}
	if msg.PkgPath == "" { // XXX
		return ErrInvalidPkgPath("missing package path")
	}
	if msg.Func == "" { // XXX
		return ErrInvalidExpr("missing function to call")
	}
	return nil
}

// Implements Msg.
func (msg MsgCall) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// Implements Msg.
func (msg MsgCall) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Caller}
}

// Implements ReceiveMsg.
func (msg MsgCall) GetReceived() std.Coins {
	return msg.Send
}

//----------------------------------------
// MsgExec

// MsgExec - executes arbitrary Gno code.
type MsgExec struct {
	Caller crypto.Address `json:"caller" yaml:"caller"`
	Send   std.Coins      `json:"send" yaml:"send"`
	Source string         `json:"source" yaml:"source"`
}

var _ std.Msg = MsgExec{}

func NewMsgExec(caller crypto.Address, send std.Coins, source string) MsgExec {
	return MsgExec{
		Caller: caller,
		Send:   send,
		Source: source,
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
	if msg.Source == "" { // XXX
		return ErrInvalidExpr("missing source to exec")
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
