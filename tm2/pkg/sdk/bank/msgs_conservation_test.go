package bank

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// A multisend must move exactly what it takes. InputOutputCoins debits every
// input and then credits every output, so the inputs-equal-outputs check is the
// only thing between a caller and coins that never existed: without it,
// {in: 100, out: 500} succeeds and leaves 500 held where 100 was.
//
// Checked at both gates. ValidateBasic is what the message router runs, and
// InputOutputCoins re-checks because it is also reachable directly.
func TestMultiSendMovesExactlyWhatItTakes(t *testing.T) {
	t.Parallel()

	a := crypto.AddressFromPreimage([]byte("conserve-a"))
	b := crypto.AddressFromPreimage([]byte("conserve-b"))
	coins := func(n int64) std.Coins { return std.NewCoins(std.NewCoin("ugnot", n)) }

	cases := []struct {
		name         string
		in, out      int64
		wantRejected bool
	}{
		{"outputs exceed inputs, which would mint", 100, 500, true},
		{"inputs exceed outputs, which would burn", 500, 100, true},
		{"equal, the only legal shape", 100, 100, false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msg := MsgMultiSend{
				Inputs:  []Input{{Address: a, Coins: coins(tt.in)}},
				Outputs: []Output{{Address: b, Coins: coins(tt.out)}},
			}
			err := msg.ValidateBasic()
			if tt.wantRejected {
				require.Error(t, err, "the router must refuse an unbalanced multisend")
				require.Contains(t, err.Error(), "sum inputs != sum outputs",
					"the refusal must name the mismatch, not some other fault")
			} else {
				require.NoError(t, err)
			}

			// The keeper must refuse it too, and move nothing when it does.
			env := setupTestEnv()
			ctx := env.ctx
			for _, ad := range []crypto.Address{a, b} {
				env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, ad))
			}
			require.NoError(t, env.bankk.MintCoins(ctx, a, coins(500)))
			held := func() int64 {
				return env.bankk.GetCoin(ctx, a, "ugnot") + env.bankk.GetCoin(ctx, b, "ugnot")
			}
			before := held()

			err = env.bankk.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
			if tt.wantRejected {
				require.Error(t, err, "the keeper must refuse an unbalanced multisend")
				require.Equal(t, before, held(), "a refused multisend must move nothing")
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, before, held(), "a multisend must never change the total held")
			require.Equal(t, int64(500), env.bankk.TotalSupply(ctx, "ugnot"),
				"a multisend must never change supply")
		})
	}
}
