package dev_test

import (
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/contribs/gnodev/pkg/dev"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePackageModifierQuery(t *testing.T) {
	validAddr := crypto.MustAddressFromString(integration.DefaultAccount_Address)
	validBech32Addr := validAddr.String()
	validCoins := std.MustParseCoins("100ugnot")

	tests := []struct {
		name       string
		path       string
		book       *address.Book
		wantQuery  dev.QueryPath
		wantErrMsg string
	}{
		{
			name: "valid creator bech32",
			path: "abc.xy/some/path?creator=" + validBech32Addr,
			book: address.NewBook(),
			wantQuery: dev.QueryPath{
				Path:    "abc.xy/some/path",
				Creator: validAddr,
			},
		},

		{
			name: "valid creator name",
			path: "abc.xy/path?creator=alice",
			book: func() *address.Book {
				bk := address.NewBook()
				bk.Add(validAddr, "alice")
				return bk
			}(),
			wantQuery: dev.QueryPath{
				Path:    "abc.xy/path",
				Creator: validAddr,
			},
		},

		{
			name:       "invalid creator",
			path:       "abc.xy/path?creator=bob",
			book:       address.NewBook(),
			wantErrMsg: `invalid name or address for creator "bob"`,
		},

		{
			name:       "invalid bech32 creator",
			path:       "abc.xy/path?creator=invalid",
			book:       address.NewBook(),
			wantErrMsg: `invalid name or address for creator "invalid"`,
		},

		{
			name: "valid deposit",
			path: "abc.xy/path?deposit=100ugnot",
			book: address.NewBook(),
			wantQuery: dev.QueryPath{
				Path:    "abc.xy/path",
				Deposit: validCoins,
			},
		},

		{
			name:       "invalid deposit",
			path:       "abc.xy/path?deposit=invalid",
			book:       address.NewBook(),
			wantErrMsg: `unable to parse deposit amount "invalid" (should be in the form xxxugnot)`,
		},

		{
			name: "both creator and deposit",
			path: "abc.xy/path?creator=" + validBech32Addr + "&deposit=100ugnot",
			book: address.NewBook(),
			wantQuery: dev.QueryPath{
				Path:    "abc.xy/path",
				Creator: validAddr,
				Deposit: validCoins,
			},
		},

		{
			name:       "malformed path",
			path:       "://invalid",
			book:       address.NewBook(),
			wantErrMsg: "malformed path/query",
		},

		{
			name: "no creator or deposit",
			path: "abc.xy/path",
			book: address.NewBook(),
			wantQuery: dev.QueryPath{
				Path: "abc.xy/path",
			},
		},

		{
			name: "clean path with ..",
			path: "abc.xy/foo/../bar",
			book: address.NewBook(),
			wantQuery: dev.QueryPath{
				Path: "abc.xy/bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, err := dev.ResolveQueryPath(tt.book, tt.path)
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantQuery, gotQuery)
		})
	}
}
