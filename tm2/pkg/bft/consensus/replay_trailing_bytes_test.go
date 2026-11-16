package consensus

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// paddedPartSet chunks block's sized encoding followed by one filler byte.
func paddedPartSet(block *types.Block) *types.PartSet {
	return types.NewPartSetFromData(append(amino.MustMarshalSized(block), 0x00), types.BlockPartSizeBytes)
}

// trailingFieldPartSet chunks block's encoding followed by an empty unknown
// field 4, inside the length prefix: the decode fills the block, then fails.
func trailingFieldPartSet(block *types.Block) *types.PartSet {
	bz := append(amino.MustMarshal(block), 0x22, 0x00)
	return types.NewPartSetFromData(append(binary.AppendUvarint(nil, uint64(len(bz))), bz...), types.BlockPartSizeBytes)
}

// TestHandshakeReplaysBlockWithTrailingPartBytes checks that a block stored
// with bytes after its length-prefixed encoding, or with an unknown field after
// its last one, as +2/3 of voting power can commit, still replays on startup
// instead of panicking.
func TestHandshakeReplaysBlockWithTrailingPartBytes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		parts func(*types.Block) *types.PartSet
	}{
		{"padded", paddedPartSet},
		{"trailing field", trailingFieldPartSet},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, genesisFile := ResetConfig("trailing_bytes_handshake_test_")
			defer os.RemoveAll(config.RootDir)

			stateDB, state, _ := makeStateAndStore(config, genesisFile, kvstore.AppVersion)
			genDoc, err := sm.MakeGenesisDocFromFile(genesisFile)
			require.NoError(t, err)

			block, _ := makeBlock(state, nil, nil, nil, 1)

			blockStore := store.NewBlockStore(memdb.NewMemDB())
			seenCommit := types.NewCommit(types.BlockID{}, []*types.CommitSig{{Height: 1, Timestamp: tmtime.Now()}})
			blockStore.SaveBlock(block, tc.parts(block), seenCommit)
			require.Equal(t, int64(1), blockStore.Height())

			// State at height 1, app write lost (fresh app at 0): handshake must replay block 1.
			buildApp := appconn.NewAppConns(proxy.NewLocalClientCreator(kvstore.NewKVStoreApplication()))
			require.NoError(t, buildApp.Start())
			defer buildApp.Stop()
			state1 := applyBlock(stateDB, state, block, buildApp)
			sm.SaveState(stateDB, state1)

			handshakeApp := appconn.NewAppConns(proxy.NewLocalClientCreator(kvstore.NewKVStoreApplication()))
			require.NoError(t, handshakeApp.Start())
			defer handshakeApp.Stop()

			handshaker := NewHandshaker(stateDB, state1, blockStore, genDoc)
			require.NoError(t, handshaker.Handshake(handshakeApp))
			require.Equal(t, 1, handshaker.NBlocks(), "block 1 is replayed from the store")
		})
	}
}
