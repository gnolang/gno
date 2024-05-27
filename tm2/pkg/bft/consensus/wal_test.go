package consensus

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	auto "github.com/gnolang/gno/tm2/pkg/autofile"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/mempool/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	walm "github.com/gnolang/gno/tm2/pkg/bft/wal"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// ----------------------------------------
// copied over from wal/wal_test.go

const maxTestMsgSize int64 = 64 * 1024

func makeTempWAL(t *testing.T, walChunkSize int64) (wal walm.WAL) {
	t.Helper()

	// Create WAL file.
	walFile := filepath.Join(t.TempDir(), "wal")

	// Create WAL.
	wal, err := walm.NewWAL(walFile, maxTestMsgSize, auto.GroupHeadSizeLimit(walChunkSize))
	if err != nil {
		panic(err)
	}
	err = wal.Start()
	if err != nil {
		panic(err)
	}

	t.Cleanup(func() {
		// WAL cleanup.
		wal.Stop()
		// wait for the wal to finish shutting down so we
		// can safely remove the directory
		wal.Wait()
	})

	return wal
}

// end copy from wal/wal_test.go
// ----------------------------------------

func TestWALTruncate(t *testing.T) {
	t.Parallel()

	const walChunkSize = 409610 // 4KB
	wal := makeTempWAL(t, walChunkSize)

	wal.SetLogger(log.NewTestingLogger(t))

	type grouper interface {
		Group() *auto.Group
	}

	// 60 block's size nearly 70K, greater than group's wal chunk filesize (4KB).
	// When the headBuf is full, content will flush to the filesystem.
	err := WALGenerateNBlocks(t, wal.(grouper).Group(), 60)
	require.NoError(t, err)

	time.Sleep(1 * time.Millisecond) // wait groupCheckDuration, make sure RotateFile run

	wal.FlushAndSync()

	h := int64(50)
	gr, found, err := wal.SearchForHeight(h+1, &walm.WALSearchOptions{})
	assert.NoError(t, err, "expected not to err on height %d", h)
	assert.True(t, found, "expected to find end height for %d", h)
	assert.NotNil(t, gr)
	defer gr.Close()

	dec := walm.NewWALReader(gr, maxMsgSize)
	msg, meta, err := dec.ReadMessage()
	assert.NoError(t, err, "expected to decode a message")
	rs, ok := msg.Msg.(newRoundStepInfo)
	assert.Nil(t, meta, "expected no meta")
	assert.True(t, ok, "expected message of type EventRoundState")
	assert.Equal(t, rs.Height, h+1, "wrong height")
}

// XXX: WALGenerateNBlocks and WALWithNBlocks were removed from wal_generator.go
// as they are unused.
// If you intend to use them, please move them to a separate package.

// WALGenerateNBlocks generates a consensus WAL. It does this by spinning up a
// stripped down version of node (proxy app, event bus, consensus state) with a
// persistent kvstore application and special consensus wal instance
// (heightStopWAL) and waits until numBlocks are created. If the node fails to produce given numBlocks, it returns an error.
func WALGenerateNBlocks(t *testing.T, wr io.Writer, numBlocks int) (err error) {
	t.Helper()

	config, genesisFile := getConfig(t)

	app := kvstore.NewPersistentKVStoreApplication(filepath.Join(config.DBDir(), "wal_generator"))
	defer app.Close()

	logger := log.NewNoopLogger().With("wal_generator", "wal_generator")
	logger.Info("generating WAL (last height msg excluded)", "numBlocks", numBlocks)

	// -----------
	// COPY PASTE FROM node.go WITH A FEW MODIFICATIONS
	// NOTE: we can't import node package because of circular dependency.
	// NOTE: we don't do handshake so need to set state.Version.Consensus.App directly.
	privValidatorKeyFile := config.PrivValidatorKeyFile()
	privValidatorStateFile := config.PrivValidatorStateFile()
	privValidator := privval.LoadOrGenFilePV(privValidatorKeyFile, privValidatorStateFile)
	genDoc, err := types.GenesisDocFromFile(genesisFile)
	if err != nil {
		return errors.Wrap(err, "failed to read genesis file")
	}
	blockStoreDB := memdb.NewMemDB()
	stateDB := blockStoreDB
	state, err := sm.MakeGenesisState(genDoc)
	if err != nil {
		return errors.Wrap(err, "failed to make genesis state")
	}
	state.AppVersion = kvstore.AppVersion
	sm.SaveState(stateDB, state)
	blockStore := store.NewBlockStore(blockStoreDB)

	proxyApp := appconn.NewAppConns(proxy.NewLocalClientCreator(app))
	proxyApp.SetLogger(logger.With("module", "proxy"))
	if err := proxyApp.Start(); err != nil {
		return errors.Wrap(err, "failed to start proxy app connections")
	}
	defer proxyApp.Stop()

	evsw := events.NewEventSwitch()
	evsw.SetLogger(logger.With("module", "events"))
	if err := evsw.Start(); err != nil {
		return errors.Wrap(err, "failed to start event bus")
	}
	defer evsw.Stop()
	mempool := mock.Mempool{}
	blockExec := sm.NewBlockExecutor(stateDB, log.NewNoopLogger(), proxyApp.Consensus(), mempool)
	consensusState := NewConsensusState(config.Consensus, state.Copy(), blockExec, blockStore, mempool)
	consensusState.SetLogger(logger)
	consensusState.SetEventSwitch(evsw)
	if privValidator != nil {
		consensusState.SetPrivValidator(privValidator)
	}
	// END OF COPY PASTE
	// -----------

	// set consensus wal to buffered WAL, which will write all incoming msgs to buffer
	numBlocksWritten := make(chan struct{})
	wal := newHeightStopWAL(logger, walm.NewWALWriter(wr, maxMsgSize), int64(numBlocks)+1, numBlocksWritten)
	// See wal.go OnStart().
	// Since we separate the WALWriter from the WAL, we need to
	// initialize ourself.
	wal.WriteMetaSync(walm.MetaMessage{Height: 1})
	consensusState.wal = wal

	if err := consensusState.Start(); err != nil {
		return errors.Wrap(err, "failed to start consensus state")
	}

	select {
	case <-numBlocksWritten:
		consensusState.Stop()
		return nil
	case <-time.After(2 * time.Minute):
		consensusState.Stop()
		return fmt.Errorf("waited too long for tendermint to produce %d blocks (grep logs for `wal_generator`)", numBlocks)
	}
}

// WALWithNBlocks returns a WAL content with numBlocks.
func WALWithNBlocks(t *testing.T, numBlocks int) (data []byte, err error) {
	t.Helper()

	var b bytes.Buffer
	wr := bufio.NewWriter(&b)

	if err := WALGenerateNBlocks(t, wr, numBlocks); err != nil {
		return []byte{}, err
	}

	wr.Flush()
	return b.Bytes(), nil
}
