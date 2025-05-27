package coins

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

func TestBalanceChecker(t *testing.T) {
	denom1 := "testtoken1"
	denom2 := "testtoken2"
	addr1 := testutils.TestAddress("user1")
	addr2 := testutils.TestAddress("user2")

	coinsRealm := std.NewCodeRealm("gno.land/r/gnoland/coins")
	testing.SetRealm(coinsRealm)

	testing.IssueCoins(addr1, std.NewCoins(std.NewCoin(denom1, 1000000)))
	testing.IssueCoins(addr2, std.NewCoins(std.NewCoin(denom1, 501)))

	testing.IssueCoins(addr2, std.NewCoins(std.NewCoin(denom2, 12345)))

	tests := []struct {
		name      string
		path      string
		contains  string
		wantPanic bool
	}{
		{
			name:     "homepage",
			path:     "",
			contains: "# Gno.land Coins Explorer",
		},
		// TODO: not supported yet
		// {
		// 	name:     "total supply",
		// 	path:     denom,
		// 	expected: "Balance: 1500000testtoken",
		// },

		{
			name:     "addr1's coin balance",
			path:     ufmt.Sprintf("balances/%s?coin=%s", addr1.String(), denom1),
			contains: ufmt.Sprintf("`%s` has `%d%s`", addr1.String(), 1000000, denom1),
		},
		{
			name:     "addr2's full balances",
			path:     ufmt.Sprintf("balances/%s", addr2.String()),
			contains: ufmt.Sprintf("This page shows full coin balances of `%s` at block", addr2.String()),
		},
		{
			name: "addr2's full balances",
			path: ufmt.Sprintf("balances/%s", addr2.String()),
			contains: `| testtoken1 | 501 |
| testtoken2 | 12345 |`,
		},
		{
			name:     "addr2's coin balance",
			path:     ufmt.Sprintf("balances/%s?coin=%s", addr2.String(), denom1),
			contains: ufmt.Sprintf("`%s` has `%d%s`", addr2.String(), 501, denom1),
		},
		{
			name:     "cosmos addr conversion",
			path:     "convert/cosmos1jg8mtutu9khhfwc4nxmuhcpftf0pajdh6svrgs",
			contains: ufmt.Sprintf("`%s` on Cosmos matches `%s`", "cosmos1jg8mtutu9khhfwc4nxmuhcpftf0pajdh6svrgs", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		},
		{
			name:      "invalid path",
			path:      "invalid",
			contains:  "404",
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for %s", tt.name)
					}
				}()
			}

			result := Render(tt.path)
			if !tt.wantPanic {
				if !strings.Contains(result, tt.contains) {
					t.Errorf("expected %s to contain %s", result, tt.contains)
				}
			}
		})
	}
}
