package gnolang

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
)

func TestTransactionStore(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, types.StoreOptions{})

	st := NewStore(nil, tm2Store, tm2Store)
	wrappedTm2Store := tm2Store.CacheWrap()
	txSt := st.BeginTransaction(wrappedTm2Store, wrappedTm2Store)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "hello",
		Store:   txSt,
		Output:  io.Discard,
	})
	_, pv := m.RunMemPackage(&std.MemPackage{
		Name: "hello",
		Path: "hello",
		Files: []*std.MemFile{
			{Name: "hello.gno", Body: "package hello; func main() { println(A(11)); }; type A int"},
		},
	}, true)
	m.SetActivePackage(pv)
	m.RunMain()

	// mem package should only exist in txSt
	// (check both memPackage and types - one is stored directly in the db,
	// the other uses txlog)
	assert.Nil(t, st.GetMemPackage("hello"))
	assert.NotNil(t, txSt.GetMemPackage("hello"))
	assert.PanicsWithValue(t, "unexpected type with id hello.A", func() { st.GetType("hello.A") })
	assert.NotNil(t, txSt.GetType("hello.A"))

	// use write on the stores
	txSt.Write()
	wrappedTm2Store.Write()

	// mem package should exist and be ==.
	res := st.GetMemPackage("hello")
	assert.NotNil(t, res)
	assert.Equal(t, txSt.GetMemPackage("hello"), res)
	helloA := st.GetType("hello.A")
	assert.NotNil(t, helloA)
	assert.Equal(t, txSt.GetType("hello.A"), helloA)
}

func TestTransactionStore_blockedMethods(t *testing.T) {
	// These methods should panic as they modify store settings, which should
	// only be changed in the root store.
	assert.Panics(t, func() { transactionStore{}.SetPackageGetter(nil) })
	assert.Panics(t, func() { transactionStore{}.ClearCache() })
	assert.Panics(t, func() { transactionStore{}.SetPackageInjector(nil) })
	assert.Panics(t, func() { transactionStore{}.SetNativeStore(nil) })
	assert.Panics(t, func() { transactionStore{}.SetStrictGo2GnoMapping(false) })
}
