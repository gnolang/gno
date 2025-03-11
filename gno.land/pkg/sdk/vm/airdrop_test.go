package vm

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAirdropClaim(t *testing.T) {
	addr1 := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))

	env := setupTestEnv()
	ctx := env.ctx

	airdrops := []AirdropInfo{
		{
			Address: addr1,
			Amount:  std.NewCoins(std.NewCoin("ugnot", 10000000)),
			Claimed: false,
		},
		{
			Address: addr2,
			Amount:  std.NewCoins(std.NewCoin("ugnot", 20000000)),
			Claimed: false,
		},
	}

	// initialize airdrop keeper
	adrpk := NewAirdropKeeper(env.vmk.iavlKey, env.bankk)

	// initialize genesis state
	adrpk.InitGenesis(ctx, GenesisState{
		Airdrops: airdrops,
	})

	t.Run("airdrop claim", func(t *testing.T) {
		// check eligibility
		eligible, amount := adrpk.IsEligible(ctx, addr1)
		require.True(t, eligible)
		assert.Equal(t, amount, std.NewCoins(std.NewCoin("ugnot", 10000000)))

		// claim airdrop
		// err := keeper.Claim(ctx, addr1)
		// require.NoError(t, err)

		// eligible, _ = keeper.IsEligible(ctx, addr1)
		// require.False(t, eligible)
	})
}
