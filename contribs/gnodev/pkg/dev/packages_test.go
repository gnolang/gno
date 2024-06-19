package dev

import (
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
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
		{".", PackagePath{
			Path: ".",
		}, false},
		{"/simple/path", PackagePath{
			Path: "/simple/path",
		}, false},
		{"/ambiguo/u//s/path///", PackagePath{
			Path: "/ambiguo/u/s/path",
		}, false},
		{"/path/with/creator?creator=testAccount", PackagePath{
			Path:    "/path/with/creator",
			Creator: testingAddress,
		}, false},
		{"/path/with/deposit?deposit=100ugnot", PackagePath{
			Path:    "/path/with/deposit",
			Deposit: std.MustParseCoins("100ugnot"),
		}, false},
		{".?creator=g1hr3dl82qdy84a5h3dmckh0suc7zgwm5rnns6na&deposit=100ugnot", PackagePath{
			Path:    ".",
			Creator: testingAddress,
			Deposit: std.MustParseCoins("100ugnot"),
		}, false},

		// errors cases
		{"/invalid/account?creator=UnknownAccount", PackagePath{}, true},
		{"/invalid/address?creator=zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", PackagePath{}, true},
		{"/invalid/deposit?deposit=abcd", PackagePath{}, true},
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
