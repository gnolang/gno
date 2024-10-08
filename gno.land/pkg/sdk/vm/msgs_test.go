package vm

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

func TestMsgAddPackage(t *testing.T) {
	creator := crypto.AddressFromPreimage([]byte("addr1"))
	pkgPath := "gno.land/r/namespace/test"
	files := []*std.MemFile{
		{
			Name: "test.gno",
			Body: `package test
		func Echo() string {return "hello world"}`,
		},
	}

	msg := NewMsgAddPackage(creator, pkgPath, files)

	// Validate Basic
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Check if package name is set correctly
	require.Equal(t, msg.Package.Name, "test")

	// Test invalid address
	msg.Creator = crypto.Address{}
	err = msg.ValidateBasic()
	require.Error(t, err)

	// Test invalid package path
	msg.Creator = creator
	msg.Package.Path = ""
	err = msg.ValidateBasic()
	require.Error(t, err)
}

func TestMsgCall(t *testing.T) {
	caller := crypto.AddressFromPreimage([]byte("addr1"))
	pkgPath := "gno.land/r/namespace/mypkg"
	funcName := "MyFunction"
	args := []string{"arg1", "arg2"}

	msg := NewMsgCall(caller, std.Coins{}, pkgPath, funcName, args)

	// Validate Basic
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Test invalid caller address
	msg.Caller = crypto.Address{}
	err = msg.ValidateBasic()
	require.Error(t, err)

	// Test invalid package path
	msg.Caller = caller
	msg.PkgPath = ""
	err = msg.ValidateBasic()
	require.Error(t, err)

	// Test invalid function name
	msg.PkgPath = pkgPath
	msg.Func = ""
	err = msg.ValidateBasic()
	require.Error(t, err)
}

func TestMsgRun(t *testing.T) {
	caller := crypto.AddressFromPreimage([]byte("addr1"))
	files := []*std.MemFile{
		{
			Name: "test.gno",
			Body: `package main
		func Echo() string {return "hello world"}`,
		},
	}

	msg := NewMsgRun(caller, std.Coins{}, files)

	// Validate Basic
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Test invalid caller address
	msg.Caller = crypto.Address{}
	err = msg.ValidateBasic()
	require.Error(t, err)

	// Test invalid package name
	files = []*std.MemFile{
		{
			Name: "test.gno",
			Body: `package test
		func Echo() string {return "hello world"}`,
		},
	}
	require.Panics(t, func() {
		NewMsgRun(caller, std.Coins{}, files)
	})
}

func TestMsgNoop(t *testing.T) {
	caller := crypto.AddressFromPreimage([]byte("addr1"))

	msg := NewMsgNoop(caller)

	// Validate Basic
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Test invalid caller address
	msg.Caller = crypto.Address{}
	err = msg.ValidateBasic()
	require.Error(t, err)
}
