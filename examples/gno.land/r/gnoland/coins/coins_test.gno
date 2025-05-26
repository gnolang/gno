package coins

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

func TestBalanceChecker(t *testing.T) {
	denom := "testtoken"
	addr1 := testutils.TestAddress("user1")
	addr2 := testutils.TestAddress("user2")

	coinsRealm := std.NewCodeRealm("gno.land/r/gnoland/coins")
	testing.SetRealm(coinsRealm)

	testing.IssueCoins(addr1, std.Coins{{denom, 1000000}})
	testing.IssueCoins(addr2, std.Coins{{denom, 500000}})

	tests := []struct {
		name      string
		path      string
		expected  string
		wantPanic bool
	}{
		{
			name:     "homepage",
			path:     "",
			expected: "# gno.land Coins Explorer",
		},
		// TODO: not supported yet
		// {
		// 	name:     "total supply",
		// 	path:     denom,
		// 	expected: "Balance: 1500000testtoken",
		// },
		{
			name:     "addr1's balance",
			path:     ufmt.Sprintf("%s/%s", denom, addr1.String()),
			expected: "Balance: 1000000testtoken",
		},
		{
			name:     "addr2's balance",
			path:     ufmt.Sprintf("%s/%s", denom, addr2.String()),
			expected: "Balance: 500000testtoken",
		},
		{
			name:      "invalid path",
			path:      ufmt.Sprintf("%s/invalid/extra", denom),
			expected:  "404",
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
				if !strings.Contains(result, tt.expected) {
					t.Errorf("expected %s to contain %s", result, tt.expected)
				}
			}
		})
	}
}
