package coins

import (
	"chain"
	"strings"
	"testing"

	"gno.land/p/leon/ctg"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/ufmt"
)

func TestBalanceChecker(t *testing.T) {
	denom1 := "testtoken1"
	denom2 := "testtoken2"
	addr1 := testutils.TestAddress("user1")
	addr2 := testutils.TestAddress("user2")

	coinsRealm := testing.NewCodeRealm("gno.land/r/gnoland/coins")
	testing.SetRealm(coinsRealm)

	testing.IssueCoins(addr1, chain.NewCoins(chain.NewCoin(denom1, 1000000)))
	testing.IssueCoins(addr2, chain.NewCoins(chain.NewCoin(denom1, 501)))

	testing.IssueCoins(addr2, chain.NewCoins(chain.NewCoin(denom2, 12345)))

	gnoAddr, _ := ctg.ConvertCosmosToGno("cosmos1s2v4tdskccx2p3yyvzem4mw5nn5fprwcku77hr")
	osmoAddr := "osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09"
	gnoAddr1, _ := ctg.ConvertAnyToGno(osmoAddr)

	testing.IssueCoins(gnoAddr1, chain.NewCoins(chain.NewCoin(denom2, 12345)))

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
			path:     ufmt.Sprintf("balances?address=%s&coin=%s", addr1.String(), denom1),
			contains: ufmt.Sprintf("`%s` has `%d%s`", addr1.String(), 1000000, denom1),
		},
		{
			name:     "addr2's full balances",
			path:     ufmt.Sprintf("balances?address=%s", addr2.String()),
			contains: ufmt.Sprintf("This page shows full coin balances of `%s` at block", addr2.String()),
		},
		{
			name: "addr2's full balances",
			path: ufmt.Sprintf("balances?address=%s", addr2.String()),
			contains: `| testtoken1 | 501 |
| testtoken2 | 12345 |`,
		},
		{
			name:     "addr2's coin balance",
			path:     ufmt.Sprintf("balances?address=%s&coin=%s", addr2.String(), denom1),
			contains: ufmt.Sprintf("`%s` has `%d%s`", addr2.String(), 501, denom1),
		},
		{
			name:     "cosmos addr conversion",
			path:     "convert/cosmos1s2v4tdskccx2p3yyvzem4mw5nn5fprwcku77hr",
			contains: ufmt.Sprintf("`cosmos1s2v4tdskccx2p3yyvzem4mw5nn5fprwcku77hr` on Cosmos matches `%s`", gnoAddr),
		},
		{
			name:     "balances bech32 auto convert",
			path:     ufmt.Sprintf("balances?address=%s&coin=%s", osmoAddr, denom1),
			contains: "Automatically converted `osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09`",
		},
		{
			name:     "single coin balance bech32 auto convert",
			path:     ufmt.Sprintf("balances?address=%s", osmoAddr),
			contains: "Automatically converted `osmo1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3aq6l09`",
		},
		{
			name:      "no addr",
			path:      "balances?address=",
			contains:  "Please input a valid address",
			wantPanic: false,
		},
		{
			name:      "no addr",
			path:      "balances?address=&coin=",
			contains:  "Please input a valid address and coin denomination.",
			wantPanic: false,
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
