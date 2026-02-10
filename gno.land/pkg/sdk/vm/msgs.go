package vm

import (
	"fmt"
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
	Creator    crypto.Address  `json:"creator" yaml:"creator"`
	Package    *std.MemPackage `json:"package" yaml:"package"`
	Send       std.Coins       `json:"send" yaml:"send"`
	MaxDeposit std.Coins       `json:"max_deposit" yaml:"max_deposit"`
}

var _ std.Msg = MsgAddPackage{}

// NewMsgAddPackage - upload a package with files.
func NewMsgAddPackage(creator crypto.Address, pkgPath string, files []*std.MemFile) MsgAddPackage {
	var pkgName string
	for _, file := range files {
		if strings.HasSuffix(file.Name, ".gno") {
			pkgName = string(gno.MustPackageNameFromFileBody(file.Name, file.Body))
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
	if !msg.Send.IsValid() {
		return std.ErrInvalidCoins(msg.Send.String())
	}
	if !msg.MaxDeposit.IsValid() {
		return std.ErrInvalidCoins(msg.MaxDeposit.String())
	}
	// Validate: ensure the package contains at least one file.
	if len(msg.Package.Files) == 0 {
		return ErrInvalidFile("no files in MsgAddPackage")
	}
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
	return msg.Send
}

//----------------------------------------
// MsgCall

// MsgCall - executes a Gno statement.
type MsgCall struct {
	Caller     crypto.Address `json:"caller" yaml:"caller"`
	Send       std.Coins      `json:"send" yaml:"send"`
	MaxDeposit std.Coins      `json:"max_deposit" yaml:"max_deposit"`
	PkgPath    string         `json:"pkg_path" yaml:"pkg_path"`
	Func       string         `json:"func" yaml:"func"`
	Args       []string       `json:"args,omitempty" yaml:"args"`
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
	if msg.PkgPath == "" {
		return ErrInvalidPkgPath("missing package path")
	}
	if !gno.IsRealmPath(msg.PkgPath) {
		return ErrInvalidPkgPath("pkgpath must be of a realm")
	}
	if _, isInt := gno.IsInternalPath(msg.PkgPath); isInt {
		return ErrInvalidPkgPath("pkgpath must not be of an internal package")
	}
	if msg.Func == "" { // XXX
		return ErrInvalidExpr("missing function to call")
	}
	if !msg.Send.IsValid() {
		return std.ErrInvalidCoins(msg.Send.String())
	}
	if !msg.MaxDeposit.IsValid() {
		return std.ErrInvalidCoins(msg.MaxDeposit.String())
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
// MsgRun

// MsgRun - executes arbitrary Gno code.
type MsgRun struct {
	Caller     crypto.Address  `json:"caller" yaml:"caller"`
	Send       std.Coins       `json:"send" yaml:"send"`
	MaxDeposit std.Coins       `json:"max_deposit" yaml:"max_deposit"`
	Package    *std.MemPackage `json:"package" yaml:"package"`
}

var _ std.Msg = MsgRun{}

func NewMsgRun(caller crypto.Address, send std.Coins, files []*std.MemFile) MsgRun {
	for _, file := range files {
		if strings.HasSuffix(file.Name, ".gno") {
			pkgName := string(gno.MustPackageNameFromFileBody(file.Name, file.Body))
			if pkgName != "main" {
				panic("package name should be 'main'")
			}
		}
	}
	return MsgRun{
		Caller: caller,
		Send:   send,
		Package: &std.MemPackage{
			Name:  "main",
			Path:  "", // auto-set by handler to fmt.Sprintf("gno.land/e/%v/run", caller.String()),
			Files: files,
		},
	}
}

// Implements Msg.
func (msg MsgRun) Route() string { return RouterKey }

// Implements Msg.
func (msg MsgRun) Type() string { return "run" }

// Implements Msg.
func (msg MsgRun) ValidateBasic() error {
	if msg.Caller.IsZero() {
		return std.ErrInvalidAddress("missing caller address")
	}

	if msg.Package.Path != "" {
		// Force memPkg path to the reserved run path.
		expected := "gno.land/e/" + msg.Caller.String() + "/run"
		if path := msg.Package.Path; path != expected {
			return ErrInvalidPkgPath(fmt.Sprintf("invalid pkgpath for MsgRun: %q", path))
		}
	}
	// Validate: ensure the package contains at least one file.
	if len(msg.Package.Files) == 0 {
		return ErrInvalidFile("no files in MsgRun")
	}

	if !msg.Send.IsValid() {
		return std.ErrInvalidCoins(msg.Send.String())
	}
	if !msg.MaxDeposit.IsValid() {
		return std.ErrInvalidCoins(msg.MaxDeposit.String())
	}
	return nil
}

// Implements Msg.
func (msg MsgRun) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// Implements Msg.
func (msg MsgRun) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Caller}
}

// Implements ReceiveMsg.
func (msg MsgRun) GetReceived() std.Coins {
	return msg.Send
}
