package immut

import (
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	iavlstore "github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// TestNewDepthEstimatorForwarding verifies that immut.New only implements
// DepthEstimator when its parent does. This prevents FixedGetReadDepth100
// (set by the VM params on every VM-context read) from overriding the flat
// ReadCostFlat rate of dbadapter-backed stores during gas simulation, which
// would cause 3× overestimation relative to the delivery path.
func TestNewDepthEstimatorForwarding(t *testing.T) {
	t.Run("iavl parent exposes DepthEstimator", func(t *testing.T) {
		db := memdb.NewMemDB()
		prefixDB := dbm.NewPrefixDB(db, []byte("s/k:main/"))
		iavlStore := iavlstore.StoreConstructor(prefixDB, types.StoreOptions{})
		wrapped := New(iavlStore)
		_, ok := wrapped.(types.DepthEstimator)
		require.True(t, ok, "immut.New(iavlStore) must implement DepthEstimator")
	})

	t.Run("dbadapter parent does not expose DepthEstimator", func(t *testing.T) {
		db := memdb.NewMemDB()
		flatStore := dbadapter.Store{DB: db}
		wrapped := New(flatStore)
		_, ok := wrapped.(types.DepthEstimator)
		require.False(t, ok, "immut.New(dbadapterStore) must NOT implement DepthEstimator")
	})

	t.Run("iavl DepthEstimator values match parent", func(t *testing.T) {
		db := memdb.NewMemDB()
		prefixDB := dbm.NewPrefixDB(db, []byte("s/k:main/"))
		iavlStore := iavlstore.StoreConstructor(prefixDB, types.StoreOptions{})
		de := iavlStore.(types.DepthEstimator)

		wrapped := New(iavlStore)
		wde := wrapped.(types.DepthEstimator)

		require.Equal(t, de.ExpectedGetReadDepth100(), wde.ExpectedGetReadDepth100())
		require.Equal(t, de.ExpectedSetReadDepth100(), wde.ExpectedSetReadDepth100())
		require.Equal(t, de.ExpectedWriteDepth100(), wde.ExpectedWriteDepth100())
	})
}
