package gas_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/gas"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func newGasKVStore() types.Store {
	meter := types.NewGasMeter(10000)
	mem := dbadapter.Store{dbm.NewMemDB()}
	return gas.New(mem, meter, types.DefaultGasConfig())
}

func bz(s string) []byte { return []byte(s) }

func keyFmt(i int) []byte { return bz(fmt.Sprintf("key%0.8d", i)) }
func valFmt(i int) []byte { return bz(fmt.Sprintf("value%0.8d", i)) }

func TestGasKVStoreBasic(t *testing.T) {
	mem := dbadapter.Store{dbm.NewMemDB()}
	meter := types.NewGasMeter(10000)
	st := gas.New(mem, meter, types.DefaultGasConfig())
	v, err := st.Get(keyFmt(1))
	require.NoError(t, err)
	require.Empty(t, v, "Expected `key1` to be empty")
	err = st.Set(keyFmt(1), valFmt(1))
	require.NoError(t, err)

	v, err = st.Get(keyFmt(1))
	require.NoError(t, err)
	require.Equal(t, valFmt(1), v)
	err = st.Delete(keyFmt(1))
	require.NoError(t, err)
	v, err = st.Get(keyFmt(1))
	require.NoError(t, err)
	require.Empty(t, v, "Expected `key1` to be empty")
	require.Equal(t, meter.GasConsumed(), types.Gas(6429))
}

func TestGasKVStoreIterator(t *testing.T) {
	mem := dbadapter.Store{dbm.NewMemDB()}
	meter := types.NewGasMeter(10000)
	st := gas.New(mem, meter, types.DefaultGasConfig())

	v1, err := st.Get(keyFmt(1))
	require.NoError(t, err)

	v2, err := st.Get(keyFmt(2))
	require.NoError(t, err)

	require.Empty(t, v1, "Expected `key1` to be empty")
	require.Empty(t, v2, "Expected `key2` to be empty")
	st.Set(keyFmt(1), valFmt(1))
	st.Set(keyFmt(2), valFmt(2))
	iterator, err := st.Iterator(nil, nil)
	require.NoError(t, err)

	ka := iterator.Key()
	require.Equal(t, ka, keyFmt(1))
	va := iterator.Value()
	require.Equal(t, va, valFmt(1))
	iterator.Next()
	kb := iterator.Key()
	require.Equal(t, kb, keyFmt(2))
	vb := iterator.Value()
	require.Equal(t, vb, valFmt(2))
	iterator.Next()
	require.False(t, iterator.Valid())
	require.Panics(t, iterator.Next)
	require.Equal(t, meter.GasConsumed(), types.Gas(6987))
}

func TestGasKVStoreOutOfGasSet(t *testing.T) {
	mem := dbadapter.Store{dbm.NewMemDB()}
	meter := types.NewGasMeter(0)
	st := gas.New(mem, meter, types.DefaultGasConfig())
	require.Panics(t, func() { st.Set(keyFmt(1), valFmt(1)) }, "Expected out-of-gas")
}

func TestGasKVStoreOutOfGasIterator(t *testing.T) {
	mem := dbadapter.Store{dbm.NewMemDB()}
	meter := types.NewGasMeter(20000)
	st := gas.New(mem, meter, types.DefaultGasConfig())
	st.Set(keyFmt(1), valFmt(1))
	iterator, err := st.Iterator(nil, nil)
	require.NoError(t, err)

	iterator.Next()
	require.Panics(t, func() { iterator.Value() }, "Expected out-of-gas")
}
