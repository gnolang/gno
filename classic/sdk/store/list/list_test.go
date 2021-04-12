package list

import (
	"math/rand"
	"testing"

	dbm "github.com/tendermint/classic/db"
	"github.com/tendermint/classic/libs/log"
	"github.com/tendermint/go-amino-x"

	abci "github.com/tendermint/classic/abci/types"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/classic/sdk/store/rootmulti"
	sdk "github.com/tendermint/classic/sdk/types"
)

type TestStruct struct {
	I uint64
	B bool
}

func defaultComponents(key sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := rootmulti.NewStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	return ctx
}
func TestList(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx := defaultComponents(key)
	store := ctx.KVStore(key)
	lm := NewList(store)

	val := TestStruct{1, true}
	var res TestStruct

	lm.Push(val)
	require.Equal(t, uint64(1), lm.Len())
	lm.Get(uint64(0), &res)
	require.Equal(t, val, res)

	val = TestStruct{2, false}
	lm.Set(uint64(0), val)
	lm.Get(uint64(0), &res)
	require.Equal(t, val, res)

	val = TestStruct{100, false}
	lm.Push(val)
	require.Equal(t, uint64(2), lm.Len())
	lm.Get(uint64(1), &res)
	require.Equal(t, val, res)

	lm.Delete(uint64(1))
	require.Equal(t, uint64(2), lm.Len())

	lm.Iterate(&res, func(index uint64) (brk bool) {
		var temp TestStruct
		lm.Get(index, &temp)
		require.Equal(t, temp, res)

		require.True(t, index != 1)
		return
	})

	lm.Iterate(&res, func(index uint64) (brk bool) {
		lm.Set(index, TestStruct{res.I + 1, !res.B})
		return
	})

	lm.Get(uint64(0), &res)
	require.Equal(t, TestStruct{3, true}, res)
}

func TestListRandom(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx := defaultComponents(key)
	store := ctx.KVStore(key)
	list := NewList(store)
	mocklist := []uint32{}

	for i := 0; i < 100; i++ {
		item := rand.Uint32()
		list.Push(item)
		mocklist = append(mocklist, item)
	}

	for k, v := range mocklist {
		var i uint32
		require.NotPanics(t, func() { list.Get(uint64(k), &i) })
		require.Equal(t, v, i)
	}
}
