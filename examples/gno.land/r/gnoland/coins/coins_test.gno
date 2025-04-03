package coins

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestBalanceChecker(t *testing.T) {
	addr1 := testutils.TestAddress("addr1")
	addr2 := testutils.TestAddress("addr2")

	realm := std.CurrentRealm()
	testDenom := realm.CoinDenom("test")
	test2Denom := realm.CoinDenom("test2")

	tests := []struct {
		name      string
		path      string
		setupFunc func()
		expected  string
		wantPanic bool
	}{
		{
			name: "homepage shows supported tokens",
			path: "",
			setupFunc: func() {
				banker := std.NewBanker(std.BankerTypeRealmIssue)
				banker.IssueCoin(addr1, test2Denom, 1000)
				banker.IssueCoin(addr2, test2Denom, 1000)
			},
			expected: `# Gno.land Coins`,
		},
		{
			name: "show total supply for test token",
			path: "/test",
			setupFunc: func() {
				banker := std.NewBanker(std.BankerTypeRealmIssue)
				banker.IssueCoin(addr1, testDenom, 1000000)
				banker.IssueCoin(addr2, testDenom, 500000)
			},
			expected: `1500000test`,
		},
	}

	for _, tt := range tests {
		if tt.setupFunc != nil {
			tt.setupFunc()
		}

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
