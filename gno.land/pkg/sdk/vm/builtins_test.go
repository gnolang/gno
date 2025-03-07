package vm

import (
	"testing"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/stretchr/testify/require"
	
	vm_pkg "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

func setupTestEnv() (sdk.Context, *vm_pkg.VMKeeper, auth.AccountKeeper, bank.BankKeeper) {
	db := memdb.NewMemDB()

	authCapKey := store.NewStoreKey("authCapKey")
	baseKey := store.NewStoreKey("baseKey")
	iavlKey := store.NewStoreKey("iavlKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, iavl.StoreConstructor, db)
	ms.MountStoreWithDB(baseKey, iavl.StoreConstructor, db)
	ms.MountStoreWithDB(iavlKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()
	
	paramk := params.NewParamsKeeper(authCapKey, "")
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "test-chain-id"}, log.NewNoopLogger())
	
	acck := auth.NewAccountKeeper(
		authCapKey, paramk, std.ProtoBaseAccount,
	)

	bankKeeper := bank.NewBankKeeper(acck)
	vmKeeper := vm_pkg.NewVMKeeper(baseKey, iavlKey, acck, bankKeeper, paramk)

	return ctx, vmKeeper, acck, bankKeeper
}

func TestSDKBanker_TotalCoin(t *testing.T) {
	ctx, vmKeeper, acck, bankKeeper := setupTestEnv()
	
	banker := vm_pkg.NewSDKBanker(vmKeeper, ctx)
	
	addr1 := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	addr3 := crypto.AddressFromPreimage([]byte("addr3"))
	
	acc1 := acck.NewAccountWithAddress(ctx, addr1)
	acc2 := acck.NewAccountWithAddress(ctx, addr2)
	acc3 := acck.NewAccountWithAddress(ctx, addr3)
	
	acck.SetAccount(ctx, acc1)
	acck.SetAccount(ctx, acc2)
	acck.SetAccount(ctx, acc3)
	
	totalGnot := banker.TotalCoin("gnot")
	require.Equal(t, int64(0), totalGnot)
	
	bankKeeper.SetCoins(ctx, addr1, std.NewCoins(std.NewCoin("gnot", 100), std.NewCoin("atom", 50)))
	bankKeeper.SetCoins(ctx, addr2, std.NewCoins(std.NewCoin("gnot", 200)))
	bankKeeper.SetCoins(ctx, addr3, std.NewCoins(std.NewCoin("atom", 75)))
	
	// existant coin
	totalGnot = banker.TotalCoin("gnot")
	require.Equal(t, int64(300), totalGnot)
		
	// non-existant coin
	totalBtc := banker.TotalCoin("btc")
	require.Equal(t, int64(0), totalBtc)
	
	bankKeeper.AddCoins(ctx, addr3, std.NewCoins(std.NewCoin("gnot", 150)))
	
	totalGnot = banker.TotalCoin("gnot")
	require.Equal(t, int64(450), totalGnot)
	
	b32addr1 := crypto.Bech32Address(addr1.String())
	banker.IssueCoin(b32addr1, "gnot", 50)
	
	totalGnot = banker.TotalCoin("gnot")
	require.Equal(t, int64(500), totalGnot)
	
	b32addr2 := crypto.Bech32Address(addr2.String())
	banker.RemoveCoin(b32addr2, "gnot", 100)
	
	totalGnot = banker.TotalCoin("gnot")
	require.Equal(t, int64(400), totalGnot)
}

func TestSDKBanker_GetCoins(t *testing.T) {
	ctx, vmKeeper, acck, bankKeeper := setupTestEnv()
	
	banker := vm_pkg.NewSDKBanker(vmKeeper, ctx)
	
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := acck.NewAccountWithAddress(ctx, addr)
	acck.SetAccount(ctx, acc)
	
	coins := banker.GetCoins(crypto.Bech32Address(addr.String()))
	require.True(t, coins.IsZero())
	
	bankKeeper.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("gnot", 100), std.NewCoin("atom", 50)))
	
	coins = banker.GetCoins(crypto.Bech32Address(addr.String()))
	require.True(t, coins.IsEqual(std.NewCoins(std.NewCoin("atom", 50), std.NewCoin("gnot", 100))))
}

func TestSDKBanker_SendCoins(t *testing.T) {
	ctx, vmKeeper, acck, bankKeeper := setupTestEnv()
	
	banker := vm_pkg.NewSDKBanker(vmKeeper, ctx)
	
	addr1 := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	
	acc1 := acck.NewAccountWithAddress(ctx, addr1)
	acc2 := acck.NewAccountWithAddress(ctx, addr2)
	acck.SetAccount(ctx, acc1)
	acck.SetAccount(ctx, acc2)
	
	bankKeeper.SetCoins(ctx, addr1, std.NewCoins(std.NewCoin("gnot", 100)))
	
	b32addr1 := crypto.Bech32Address(addr1.String())
	b32addr2 := crypto.Bech32Address(addr2.String())
	banker.SendCoins(b32addr1, b32addr2, std.NewCoins(std.NewCoin("gnot", 50)))
	
	coins1 := banker.GetCoins(b32addr1)
	coins2 := banker.GetCoins(b32addr2)
	
	require.True(t, coins1.IsEqual(std.NewCoins(std.NewCoin("gnot", 50))))
	require.True(t, coins2.IsEqual(std.NewCoins(std.NewCoin("gnot", 50))))
	
	require.Panics(t, func() {
		banker.SendCoins(b32addr1, b32addr2, std.NewCoins(std.NewCoin("gnot", 100)))
	})
}

func TestSDKBanker_IssueCoin(t *testing.T) {
	ctx, vmKeeper, acck, _ := setupTestEnv()
	
	banker := vm_pkg.NewSDKBanker(vmKeeper, ctx)
	
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := acck.NewAccountWithAddress(ctx, addr)
	acck.SetAccount(ctx, acc)
	
	b32addr := crypto.Bech32Address(addr.String())
	banker.IssueCoin(b32addr, "gnot", 100)
	
	coins := banker.GetCoins(b32addr)
	require.True(t, coins.IsEqual(std.NewCoins(std.NewCoin("gnot", 100))))
	
	banker.IssueCoin(b32addr, "atom", 50)
	
	coins = banker.GetCoins(b32addr)
	require.True(t, coins.IsEqual(std.NewCoins(std.NewCoin("atom", 50), std.NewCoin("gnot", 100))))
}

func TestSDKBanker_RemoveCoin(t *testing.T) {
	ctx, vmKeeper, acck, bankKeeper := setupTestEnv()
	
	banker := vm_pkg.NewSDKBanker(vmKeeper, ctx)
	
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := acck.NewAccountWithAddress(ctx, addr)
	acck.SetAccount(ctx, acc)
	
	bankKeeper.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("gnot", 100), std.NewCoin("atom", 50)))
	
	b32addr := crypto.Bech32Address(addr.String())
	banker.RemoveCoin(b32addr, "gnot", 30)
	
	coins := banker.GetCoins(b32addr)
	require.True(t, coins.IsEqual(std.NewCoins(std.NewCoin("atom", 50), std.NewCoin("gnot", 70))))
	
	require.Panics(t, func() {
		banker.RemoveCoin(b32addr, "gnot", 100)
	})
} 
