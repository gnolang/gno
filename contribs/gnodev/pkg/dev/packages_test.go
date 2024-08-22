package dev

import (
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePackagePathQuery(t *testing.T) {
	t.Parallel()

	var (
		testingName    = "testAccount"
		testingAddress = crypto.MustAddressFromString("g1hr3dl82qdy84a5h3dmckh0suc7zgwm5rnns6na")
	)

	book := address.NewBook()
	book.Add(testingAddress, testingName)

	cases := []struct {
		Path                string
		ExpectedPackagePath PackagePath
		ShouldFail          bool
	}{
		{
			Path: ".",
			ExpectedPackagePath: PackagePath{
				Path: ".",
			},
		},
		{
			Path: "/simple/path",
			ExpectedPackagePath: PackagePath{
				Path: "/simple/path",
			},
		},
		{
			Path: "/ambiguo/u//s/path///",
			ExpectedPackagePath: PackagePath{
				Path: "/ambiguo/u/s/path",
			},
		},
		{
			Path: "/path/with/creator?creator=testAccount",
			ExpectedPackagePath: PackagePath{
				Path:    "/path/with/creator",
				Creator: testingAddress,
			},
		},
		{
			Path: "/path/with/deposit?deposit=" + ugnot.ValueString(100),
			ExpectedPackagePath: PackagePath{
				Path:    "/path/with/deposit",
				Deposit: std.MustParseCoins(ugnot.ValueString(100)),
			},
		},
		{
			Path: ".?creator=g1hr3dl82qdy84a5h3dmckh0suc7zgwm5rnns6na&deposit=" + ugnot.ValueString(100),
			ExpectedPackagePath: PackagePath{
				Path:    ".",
				Creator: testingAddress,
				Deposit: std.MustParseCoins(ugnot.ValueString(100)),
			},
		},

		// errors cases
		{
			Path:       "/invalid/account?creator=UnknownAccount",
			ShouldFail: true,
		},
		{
			Path:       "/invalid/address?creator=zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			ShouldFail: true,
		},
		{
			Path:       "/invalid/deposit?deposit=abcd",
			ShouldFail: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Path, func(t *testing.T) {
			t.Parallel()

			result, err := ResolvePackagePathQuery(book, tc.Path)
			if tc.ShouldFail {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.ExpectedPackagePath.Path, result.Path)
			assert.Equal(t, tc.ExpectedPackagePath.Creator, result.Creator)
			assert.Equal(t, tc.ExpectedPackagePath.Deposit.String(), result.Deposit.String())
		})
	}
}
