package state

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestLoadABCIResponsesOfHeldEmptyBlock pins that an empty block's results,
// which encode to no bytes, read back while block sync holds them in memory,
// as RPC reads them through the node's state database.
func TestLoadABCIResponsesOfHeldEmptyBlock(t *testing.T) {
	t.Parallel()

	db := dbm.NewHoldingDB(memdb.NewMemDB())
	db.Hold(true)
	empty := &ABCIResponses{EndBlock: abci.ResponseEndBlock{}}
	require.NoError(t, SaveABCIResponses(db, 7, empty))

	loaded, err := LoadABCIResponses(db, 7)
	require.NoError(t, err, "a held empty block's results read as missing")
	require.Equal(t, empty.Bytes(), loaded.Bytes())
}
