package coins

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestAirdropChecker(t *testing.T) {
	addr1 := testutils.TestAddress("addr1")
	addr2 := testutils.TestAddress("addr2")

	realm := std.CurrentRealm()
	testDenom := realm.CoinDenom("test")

	tests := []struct {
		name      string
		path      string
		setupFunc func(std.Banker)
		expected  string
		wantPanic bool
	}{
		{
			name: "homepage shows supported tokens",
			path: "",
			setupFunc: func(banker std.Banker) {
				banker.IssueCoin(addr1, testDenom, 1000)
				banker.IssueCoin(addr2, testDenom, 1000)
			},
			expected: `# Gno.land Coins`,
		},
		// {
		// 	name: "show total supply for test token",
		// 	path: "test",
		// 	setupFunc: func(banker std.Banker) {
		// 		banker.IssueCoin(addr1, testDenom, 1000000)
		//         banker.IssueCoin(addr2, testDenom, 500000)
		// 	},
		// 	expected: `- Total Supply: 1500000test`,
		// },
	}

	for _, tt := range tests {
		banker := std.NewBanker(std.BankerTypeRealmIssue)
		tt.setupFunc(banker)

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
	}
}
