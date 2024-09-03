package vm

import (
	"fmt"
	"regexp"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var reMetaFieldName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

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
		return std.ErrInvalidCoins(msg.Deposit.String())
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
	if msg.PkgPath == "" {
		return ErrInvalidPkgPath("missing package path")
	}
	if !gno.IsRealmPath(msg.PkgPath) {
		return ErrInvalidPkgPath("pkgpath must be of a realm")
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
// MsgRun

// MsgRun - executes arbitrary Gno code.
type MsgRun struct {
	Caller  crypto.Address  `json:"caller" yaml:"caller"`
	Send    std.Coins       `json:"send" yaml:"send"`
	Package *std.MemPackage `json:"package" yaml:"package"`
}

var _ std.Msg = MsgRun{}

func NewMsgRun(caller crypto.Address, send std.Coins, files []*std.MemFile) MsgRun {
	for _, file := range files {
		if strings.HasSuffix(file.Name, ".gno") {
			pkgName := string(gno.PackageNameFromFileBody(file.Name, file.Body))
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
			Path:  "", // auto set by the handler
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

	// Force memPkg path to the reserved run path.
	wantPath := "gno.land/r/" + msg.Caller.String() + "/run"
	if path := msg.Package.Path; path != "" && path != wantPath {
		return ErrInvalidPkgPath(fmt.Sprintf("invalid pkgpath for MsgRun: %q", path))
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

//----------------------------------------
// MsgSetMeta

type (
	// MsgSetMeta - set package metadata.
	MsgSetMeta struct {
		Caller  crypto.Address `json:"caller" yaml:"caller"`
		Send    std.Coins      `json:"send" yaml:"send"`
		PkgPath string         `json:"pkg_path" yaml:"pkg_path"`
		Fields  []*MetaField   `json:"fields" yaml:"fields"`
	}

	// MetaField defines a package metadata field.
	MetaField struct {
		Name  string `json:"name" yaml:"name"`
		Value []byte `json:"value" yaml:"value"`
	}
)

var _ std.Msg = MsgSetMeta{}

func NewMsgSetMeta(caller crypto.Address, send std.Coins, pkgPath string, fields []*MetaField) MsgSetMeta {
	return MsgSetMeta{
		Caller:  caller,
		Send:    send,
		PkgPath: pkgPath,
		Fields:  fields,
	}
}

// Implements Msg.
func (msg MsgSetMeta) Route() string { return RouterKey }

// Implements Msg.
func (msg MsgSetMeta) Type() string { return "meta" }

// Implements Msg.
func (msg MsgSetMeta) ValidateBasic() error {
	if msg.Caller.IsZero() {
		return std.ErrInvalidAddress("missing caller address")
	}

	if msg.PkgPath == "" {
		return ErrInvalidPkgPath("missing package path")
	}

	if !strings.HasPrefix(msg.PkgPath, "gno.land/") {
		return ErrInvalidPkgPath("pkgpath must be of a realm or package")
	}

	count := len(msg.Fields)
	if count == 0 {
		return ErrInvalidPkgMeta("missing metadata fields")
	}

	if count > maxMetaFields {
		return ErrInvalidPkgMeta("maximum number of package metadata fields reached")
	}

	for _, f := range msg.Fields {
		if !reMetaFieldName.Match([]byte(f.Name)) {
			return ErrInvalidPkgMeta("invalid metadata field name")
		}

		if len(f.Value) > maxMetaFieldValueSize {
			return ErrInvalidPkgMeta("metadata field value is too large")
		}
	}
	return nil
}

// Implements Msg.
func (msg MsgSetMeta) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

// Implements Msg.
func (msg MsgSetMeta) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Caller}
}

// Implements ReceiveMsg.
func (msg MsgSetMeta) GetReceived() std.Coins {
	return msg.Send
}
