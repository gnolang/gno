package vm

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
)

func TestMsgAddPackage_ValidateBasic(t *testing.T) {
	t.Parallel()

	creator := crypto.AddressFromPreimage([]byte("addr1"))
	pkgName := "test"
	pkgPath := "gno.land/r/namespace/test"
	files := []*std.MemFile{
		{
			Name: "test.gno",
			Body: `package test
		func Echo() string {return "hello world"}`,
		},
	}

	tests := []struct {
		name            string
		msg             MsgAddPackage
		expectSignBytes string
		expectErr       error
	}{
		{
			name:            "valid message",
			msg:             NewMsgAddPackage(creator, pkgPath, files),
			expectSignBytes: `{"creator":"g14ch5q26mhx3jk5cxl88t278nper264ces4m8nt","max_deposit":"","package":{"files":[{"body":"package test\n\t\tfunc Echo() string {return \"hello world\"}","name":"test.gno"}],"name":"test","path":"gno.land/r/namespace/test"},"send":""}`,
			expectErr:       nil,
		},
		{
			name: "missing creator address",
			msg: MsgAddPackage{
				Creator: crypto.Address{},
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  pkgPath,
					Files: files,
				},
				MaxDeposit: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: std.InvalidAddressError{},
		},
		{
			name: "missing package path",
			msg: MsgAddPackage{
				Creator: creator,
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  "",
					Files: files,
				},
				MaxDeposit: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: InvalidPkgPathError{},
		},
		{
			name: "invalid deposit coins",
			msg: MsgAddPackage{
				Creator: creator,
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  pkgPath,
					Files: files,
				},
				MaxDeposit: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: -1000, // invalid amount
				}},
			},
			expectErr: std.InvalidCoinsError{},
		},
		{
			name: "invalid Send coins",
			msg: MsgAddPackage{
				Creator: creator,
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  pkgPath,
					Files: files,
				},
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: -1000,
				}},
			},
			expectErr: std.InvalidCoinsError{},
		},
		{
			name: "invalid MaxDeposit coins",
			msg: MsgAddPackage{
				Creator: creator,
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  pkgPath,
					Files: files,
				},
				MaxDeposit: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: -1000,
				}},
			},
			expectErr: std.InvalidCoinsError{},
		},
		{
			name: "empty files array",
			msg: MsgAddPackage{
				Creator: creator,
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  pkgPath,
					Files: []*std.MemFile{},
				},
			},
			expectErr: InvalidFileError{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.msg.ValidateBasic(); err != nil {
				assert.ErrorIs(t, err, tc.expectErr)
			} else {
				assert.Equal(t, tc.expectSignBytes, string(tc.msg.GetSignBytes()))
			}
		})
	}
}

func TestMsgCall_ValidateBasic(t *testing.T) {
	t.Parallel()

	caller := crypto.AddressFromPreimage([]byte("addr1"))
	pkgPath := "gno.land/r/namespace/test"
	funcName := "MyFunction"
	args := []string{"arg1", "arg2"}

	tests := []struct {
		name            string
		msg             MsgCall
		expectSignBytes string
		expectErr       error
	}{
		{
			name: "valid message",
			msg:  NewMsgCall(caller, std.NewCoins(std.NewCoin("ugnot", 1000)), pkgPath, funcName, args),
			expectSignBytes: `{"args":["arg1","arg2"],"caller":"g14ch5q26mhx3jk5cxl88t278nper264ces4m8nt",` +
				`"func":"MyFunction","max_deposit":"","pkg_path":"gno.land/r/namespace/test","send":"1000ugnot"}`,
			expectErr: nil,
		},
		{
			name: "invalid caller address",
			msg: MsgCall{
				Caller:  crypto.Address{},
				PkgPath: pkgPath,
				Func:    funcName,
				Args:    args,
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: std.InvalidAddressError{},
		},
		{
			name: "missing package path",
			msg: MsgCall{
				Caller:  caller,
				PkgPath: "",
				Func:    funcName,
				Args:    args,
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: InvalidPkgPathError{},
		},
		{
			name: "pkgPath should not be a realm path",
			msg: MsgCall{
				Caller:  caller,
				PkgPath: "gno.land/p/namespace/test", // this is not a valid realm path
				Func:    funcName,
				Args:    args,
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: InvalidPkgPathError{},
		},
		{
			name: "pkgPath should not be an internal path",
			msg: MsgCall{
				Caller:  caller,
				PkgPath: "gno.land/r/demo/avl/internal/sort",
				Func:    funcName,
				Args:    args,
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: InvalidPkgPathError{},
		},
		{
			name: "missing function name to call",
			msg: MsgCall{
				Caller:  caller,
				PkgPath: pkgPath,
				Func:    "",
				Args:    args,
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: InvalidExprError{},
		},
		{
			name: "invalid Send coins",
			msg: MsgCall{
				Caller:  caller,
				PkgPath: pkgPath,
				Func:    funcName,
				Args:    []string{},
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: -1000,
				}},
			},
			expectErr: std.InvalidCoinsError{},
		},
		{
			name: "invalid MaxDeposit coins",
			msg: MsgCall{
				Caller:  caller,
				PkgPath: pkgPath,
				Func:    funcName,
				Args:    []string{},
				MaxDeposit: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: -1000,
				}},
			},
			expectErr: std.InvalidCoinsError{},
		},
		{
			name: "empty arguments",
			msg: MsgCall{
				Caller:  caller,
				PkgPath: pkgPath,
				Func:    funcName,
				Args:    []string{},
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectSignBytes: `{"caller":"g14ch5q26mhx3jk5cxl88t278nper264ces4m8nt","func":"MyFunction","max_deposit":"","pkg_path":"gno.land/r/namespace/test","send":"1000ugnot"}`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.msg.ValidateBasic(); err != nil {
				assert.ErrorIs(t, err, tc.expectErr)
			} else {
				assert.Equal(t, tc.expectSignBytes, string(tc.msg.GetSignBytes()))
			}
		})
	}
}

func TestMsgRun_ValidateBasic(t *testing.T) {
	t.Parallel()

	caller := crypto.AddressFromPreimage([]byte("addr1"))
	pkgName := "main"
	pkgPath := "gno.land/e/" + caller.String() + "/run"
	pkgFiles := []*std.MemFile{
		{
			Name: "main.gno",
			Body: `package main
		func Echo() string {return "hello world"}`,
		},
	}

	tests := []struct {
		name            string
		msg             MsgRun
		expectSignBytes string
		expectErr       error
	}{
		{
			name:            "valid message",
			msg:             NewMsgRun(caller, std.NewCoins(std.NewCoin("ugnot", 1000)), pkgFiles),
			expectSignBytes: `{"caller":"g14ch5q26mhx3jk5cxl88t278nper264ces4m8nt","max_deposit":"","package":{"files":[{"body":"package main\n\t\tfunc Echo() string {return \"hello world\"}","name":"main.gno"}],"name":"main","path":""},"send":"1000ugnot"}`,
			expectErr:       nil,
		},
		{
			name: "invalid caller address",
			msg: MsgRun{
				Caller: crypto.Address{},
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  pkgPath,
					Files: pkgFiles,
				},
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: std.InvalidAddressError{},
		},
		{
			name: "invalid package path",
			msg: MsgRun{
				Caller: caller,
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  "gno.land/r/namespace/test", // this is not a valid run path
					Files: pkgFiles,
				},
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: InvalidPkgPathError{},
		},
		{
			name: "invalid Send coins",
			msg: MsgRun{
				Caller: caller,
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  pkgPath,
					Files: pkgFiles,
				},
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: -1000,
				}},
			},
			expectErr: std.InvalidCoinsError{},
		},
		{
			name: "invalid MaxDeposit coins",
			msg: MsgRun{
				Caller: caller,
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  pkgPath,
					Files: pkgFiles,
				},
				MaxDeposit: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: -1000,
				}},
			},
			expectErr: std.InvalidCoinsError{},
		},
		{
			name: "empty package files",
			msg: MsgRun{
				Caller: caller,
				Package: &std.MemPackage{
					Name:  pkgName,
					Path:  pkgPath,
					Files: []*std.MemFile{},
				},
				Send: std.Coins{std.Coin{
					Denom:  "ugnot",
					Amount: 1000,
				}},
			},
			expectErr: InvalidFileError{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if err := tc.msg.ValidateBasic(); err != nil {
				assert.ErrorIs(t, err, tc.expectErr)
			} else {
				assert.Equal(t, tc.expectSignBytes, string(tc.msg.GetSignBytes()))
			}
		})
	}
}
