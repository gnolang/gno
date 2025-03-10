package vm

import (
	_ "strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/stretchr/testify/assert"
)

func TestSDKBankerTotalCoin(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	banker := NewSDKBanker(env.vmk, ctx)

	// create test accounts and set coins
	addr1 := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))

	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	env.acck.SetAccount(ctx, acc1)
	env.bank.SetCoins(ctx, addr1, std.MustParseCoins("1000ugnot,500atom"))

	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	env.acck.SetAccount(ctx, acc2)
	env.bank.SetCoins(ctx, addr2, std.MustParseCoins("2000ugnot,1500atom"))

	tests := []struct {
		name      string
		denom     string
		expected  int64
		mustPanic bool
	}{
		{"ugnot total", "ugnot", 3000, false},
		{"atom total", "atom", 2000, false},
		{"non-existent denom", "foo", 0, false},
		{"zero balance accounts included", "ugnot", 3000, false},
		{"empty string denom", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mustPanic {
				assert.Panics(t, func() {
					banker.TotalCoin(tt.denom)
				})
			} else {
				total := banker.TotalCoin(tt.denom)
				assert.Equal(t, tt.expected, total)
			}
		})
	}
}
