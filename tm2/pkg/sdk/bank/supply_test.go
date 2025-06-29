package bank

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStore struct {
	data map[string][]byte
}

var _ types.Store = (*mockStore)(nil)

func newMockStore() *mockStore {
	return &mockStore{
		data: make(map[string][]byte),
	}
}

func (ms *mockStore) Get(key []byte) []byte {
	return ms.data[string(key)]
}

func (ms *mockStore) Has(key []byte) bool {
	_, ok := ms.data[string(key)]
	return ok
}

func (ms *mockStore) Set(key []byte, value []byte) {
	ms.data[string(key)] = value
}

func (ms *mockStore) Delete(key []byte) {
	delete(ms.data, string(key))
}

func (ms *mockStore) Iterator(start, end []byte) types.Iterator {
	panic("not implemented")
}

func (ms *mockStore) ReverseIterator(start, end []byte) types.Iterator {
	panic("not implemented")
}

func (ms *mockStore) CacheWrap() types.Store {
	panic("not implemented")
}

func (ms *mockStore) Write() {
	panic("not implemented")
}

func TestSupplyStore(t *testing.T) {
	const (
		ugnot = "ugnot"
		foo   = "foo"
		bar   = "bar"
		baz   = "baz"
	)

	t.Run("basic test", func(t *testing.T) {
		store, supplyStore := createSupplyStore(t)

		// initial supply must be 0
		supply, err := supplyStore.GetSupply(store, ugnot)
		require.NoError(t, err)
		assert.Equal(t, int64(0), supply)

		// set supply
		err = supplyStore.SetSupply(store, ugnot, 1000)
		require.NoError(t, err)

		supply, err = supplyStore.GetSupply(store, ugnot)
		require.NoError(t, err)
		assert.Equal(t, int64(1000), supply)
	})

	t.Run("increase, decrease supply amount", func(t *testing.T) {
		store, supplyStore := createSupplyStore(t)

		err := supplyStore.AddSupply(store, ugnot, 500)
		require.NoError(t, err)

		supply, err := supplyStore.GetSupply(store, ugnot)
		require.NoError(t, err)
		assert.Equal(t, int64(500), supply)

		// add more supply
		err = supplyStore.AddSupply(store, ugnot, 300)
		require.NoError(t, err)

		supply, err = supplyStore.GetSupply(store, ugnot)
		require.NoError(t, err)
		assert.Equal(t, int64(800), supply)

		// decrease supply
		err = supplyStore.SubtractSupply(store, ugnot, 300)
		require.NoError(t, err)

		supply, err = supplyStore.GetSupply(store, ugnot)
		require.NoError(t, err)
		assert.Equal(t, int64(500), supply)
	})

	t.Run("error cases", func(t *testing.T) {
		const emptyDenom = ""
		store, supplyStore := createSupplyStore(t)

		// search supply with empty denom
		_, err := supplyStore.GetSupply(store, emptyDenom)
		assert.Equal(t, errEmptyDenom, err)

		// set supply with empty denom
		err = supplyStore.SetSupply(store, emptyDenom, 1000)
		assert.Equal(t, errEmptyDenom, err)

		// not enough supply to subtract
		err = supplyStore.SubtractSupply(store, ugnot, 1000)
		assert.Equal(t, errInsufficientSupply, err)
	})

	t.Run("handle multiple denoms", func(t *testing.T) {
		store, supplyStore := createSupplyStore(t)

		denoms := []string{foo, bar, baz, ugnot}
		amounts := []int64{1000, 2000, 3000, 4000}

		for i, denom := range denoms {
			err := supplyStore.SetSupply(store, denom, amounts[i])
			require.NoError(t, err)
		}

		for i, denom := range denoms {
			supply, err := supplyStore.GetSupply(store, denom)
			require.NoError(t, err)
			assert.Equal(t, amounts[i], supply)
		}

		// change other token's supply
		err := supplyStore.AddSupply(store, ugnot, 500)
		require.NoError(t, err)

		supply, err := supplyStore.GetSupply(store, foo)
		require.NoError(t, err)
		assert.Equal(t, int64(1000), supply)

		supply, err = supplyStore.GetSupply(store, bar)
		require.NoError(t, err)
		assert.Equal(t, int64(2000), supply)
	})
}

func createSupplyStore(t *testing.T) (*mockStore, *SupplyStore) {
	t.Helper()

	store := newMockStore()
	supplyStore := NewSupplyStore(store)
	return store, supplyStore
}
