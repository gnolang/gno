package auth

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// coinsAddTotal is the pre-check sum written as one std.Coins.Add per
// message, the reference sessionSpendTotal must agree with.
func coinsAddTotal(tx std.Tx, signer crypto.Address) std.Coins {
	total := std.Coins{}
	if !tx.Fee.GasFee.IsZero() {
		total = total.Add(std.Coins{tx.Fee.GasFee})
	}
	for _, msg := range tx.GetMsgs() {
		if est, ok := msg.(std.SpendEstimator); ok {
			total = total.Add(est.SpendForSigner(signer))
		}
	}
	return total
}

// sumOrPanic returns what sum returns, or the value it panicked with.
func sumOrPanic(sum func(std.Tx, crypto.Address) std.Coins, tx std.Tx, signer crypto.Address) (total std.Coins, panicked any) {
	defer func() { panicked = recover() }()
	return sum(tx, signer), nil
}

// Over generated transactions whose coins are sorted, distinct and positive, as
// every SpendEstimator's ValidateBasic requires, the sum and the overflow
// panic, message included, match one Coins.Add per message.
func TestSessionSpendTotalMatchesCoinsAdd(t *testing.T) {
	t.Parallel()

	_, _, signer := tu.KeyTestPubAddr()
	_, _, other := tu.KeyTestPubAddr()
	denoms := []string{"atom", "btc", "eth", "gnot", "ugnot", "xyz"}
	rng := rand.New(rand.NewSource(1))
	amount := func() int64 {
		if rng.Intn(4) == 0 {
			return math.MaxInt64 - rng.Int63n(1000) // overflows once summed
		}
		return 1 + rng.Int63n(1000)
	}
	// coins returns a set ValidateBasic accepts: sorted, distinct, positive.
	coins := func() std.Coins {
		var cs std.Coins
		for _, d := range denoms {
			if rng.Intn(3) == 0 {
				cs = append(cs, std.NewCoin(d, amount()))
			}
		}
		return cs
	}

	panics := 0
	for i := range 20000 {
		var fee std.Coin
		if rng.Intn(4) != 0 {
			fee = std.NewCoin(denoms[rng.Intn(len(denoms))], amount())
		}
		msgs := make([]std.Msg, rng.Intn(8))
		for j := range msgs {
			switch rng.Intn(4) {
			case 0:
				msgs[j] = tu.NewTestMsg(signer) // declares no spend
			case 1:
				msgs[j] = tu.MockMsgSend{From: other, Amount: coins()}
			default:
				msgs[j] = tu.MockMsgSend{From: signer, Amount: coins()}
			}
		}
		tx := std.Tx{Msgs: msgs, Fee: std.Fee{GasFee: fee}}

		// The sum runs first, so one that writes into a message's coins, which
		// DeliverTx later sends, changes what the reference reads.
		got, gotPanic := sumOrPanic(sessionSpendTotal, tx, signer)
		want, wantPanic := sumOrPanic(coinsAddTotal, tx, signer)
		require.Equal(t, wantPanic, gotPanic, "case %d", i)
		require.Equal(t, want.String(), got.String(), "case %d", i)
		if wantPanic != nil {
			panics++
		}
	}
	require.Greater(t, panics, 1000, "overflow cases generated")
}

// One message per denom grows the pre-check's sum by one coin each. The
// pre-check must allocate the same per message however many came before it.
func TestSessionAntePreCheckLinearInDenoms(t *testing.T) {
	// Not parallel: MemStats also counts what concurrent tests allocate.
	env, anteHandler, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	// The limit holds only atom: the pre-check sums every message, then refuses.
	sa := createSessionDirect(t, env, masterAddr, sessionPub, 0)

	// allocated returns the bytes the ante handler allocates for n messages.
	allocated := func(n int) uint64 {
		msgs := make([]std.Msg, n)
		for i := range msgs {
			msgs[i] = tu.MockMsgSend{
				From:   masterAddr,
				Amount: std.Coins{std.NewCoin(fmt.Sprintf("d%05d", i), 1)},
			}
		}
		tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, tu.NewTestFee())

		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		_, res, abort := anteHandler(ctx, tx, false)
		runtime.ReadMemStats(&after)
		require.True(t, abort)
		require.IsType(t, std.SessionNotAllowedError{}, res.Error)
		// The pre-check refuses past its empty-limit return, naming the
		// message denoms; the fee alone fits the limit.
		require.Contains(t, res.Log, "would be exceeded")
		require.Contains(t, res.Log, "d00000")
		return after.TotalAlloc - before.TotalAlloc
	}

	small, large := allocated(500), allocated(2000)
	// Four times the messages: linear work allocates about four times as
	// much, one Coins.Add per message sixteen.
	require.Less(t, large, 8*small, "500 messages allocated %d bytes, 2000 allocated %d", small, large)
}
