package vm

import (
	"fmt"
	"github.com/google/uuid"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

//----------------------------------------
// MsgAddPackage

// MsgAddPackage - create and initialize new package
type MsgAddPackage struct {
	Creator uuid.UUID       `json:"creator" yaml:"creator"`
	Package *std.MemPackage `json:"package" yaml:"package"`
}

// NewMsgAddPackage - upload a package with files.
func NewMsgAddPackage(creator uuid.UUID, pkgPath string, files []*std.MemFile) MsgAddPackage {
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
	if msg.Package.Path == "" { // XXX
		return ErrInvalidPkgPath("missing package path")
	}
	return nil
}

// Implements Msg.
func (msg MsgAddPackage) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}

//----------------------------------------
// MsgCall

// MsgCall - executes a Gno statement.
type MsgCall struct {
	Caller  uuid.UUID `json:"caller" yaml:"caller"`
	PkgPath string    `json:"pkg_path" yaml:"pkg_path"`
	Func    string    `json:"func" yaml:"func"`
	Args    []string  `json:"args" yaml:"args"`
}

func NewMsgCall(caller uuid.UUID, send sdk.Coins, pkgPath, fnc string, args []string) MsgCall {
	return MsgCall{
		Caller:  caller,
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

//----------------------------------------
// MsgRun

// MsgRun - executes arbitrary Gno code.
type MsgRun struct {
	Caller  uuid.UUID       `json:"caller" yaml:"caller"`
	Package *std.MemPackage `json:"package" yaml:"package"`
}

func NewMsgRun(caller uuid.UUID, send std.Coins, files []*std.MemFile) MsgRun {
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
