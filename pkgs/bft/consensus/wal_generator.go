package consensus

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/pkgs/bft/abci/example/kvstore"
	cfg "github.com/gnolang/gno/pkgs/bft/config"
	"github.com/gnolang/gno/pkgs/bft/mempool/mock"
	"github.com/gnolang/gno/pkgs/bft/privval"
	"github.com/gnolang/gno/pkgs/bft/proxy"
	sm "github.com/gnolang/gno/pkgs/bft/state"
	"github.com/gnolang/gno/pkgs/bft/store"
	"github.com/gnolang/gno/pkgs/bft/types"
	walm "github.com/gnolang/gno/pkgs/bft/wal"
	db "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/events"
	"github.com/gnolang/gno/pkgs/log"
	"github.com/gnolang/gno/pkgs/random"
)

// WALGenerateNBlocks generates a consensus WAL. It does this by spinning up a
// stripped down version of node (proxy app, event bus, consensus state) with a
// persistent kvstore application and special consensus wal instance
// (heightStopWAL) and waits until numBlocks are created. If the node fails to produce given numBlocks, it returns an error.
func WALGenerateNBlocks(t *testing.T, wr io.Writer, numBlocks int) (err error) {
	t.Helper()

	config := getConfig(t)

	app := kvstore.NewPersistentKVStoreApplication(filepath.Join(config.DBDir(), "wal_generator"))
	defer app.Close()

	logger := log.TestingLogger().With("wal_generator", "wal_generator")
	logger.Info("generating WAL (last height msg excluded)", "numBlocks", numBlocks)

	/////////////////////////////////////////////////////////////////////////////
	// COPY PASTE FROM node.go WITH A FEW MODIFICATIONS
	// NOTE: we can't import node package because of circular dependency.
	// NOTE: we don't do handshake so need to set state.Version.Consensus.App directly.
	privValidatorKeyFile := config.PrivValidatorKeyFile()
	privValidatorStateFile := config.PrivValidatorStateFile()
	privValidator := privval.LoadOrGenFilePV(privValidatorKeyFile, privValidatorStateFile)
	genDoc, err := types.GenesisDocFromFile(config.GenesisFile())
	if err != nil {
		return errors.Wrap(err, "failed to read genesis file")
	}
	blockStoreDB := db.NewMemDB()
	stateDB := blockStoreDB
	state, err := sm.MakeGenesisState(genDoc)
	if err != nil {
		return errors.Wrap(err, "failed to make genesis state")
	}
	state.AppVersion = kvstore.AppVersion
	sm.SaveState(stateDB, state)
	blockStore := store.NewBlockStore(blockStoreDB)

	proxyApp := proxy.NewAppConns(proxy.NewLocalClientCreator(app))
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
	blockExec := sm.NewBlockExecutor(stateDB, log.TestingLogger(), proxyApp.Consensus(), mempool)
	consensusState := NewConsensusState(config.Consensus, state.Copy(), blockExec, blockStore, mempool)
	consensusState.SetLogger(logger)
	consensusState.SetEventSwitch(evsw)
	if privValidator != nil {
		consensusState.SetPrivValidator(privValidator)
	}
	// END OF COPY PASTE
	/////////////////////////////////////////////////////////////////////////////

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

func randPort() int {
	// returns between base and base + spread
	base, spread := 20000, 20000

	return base + random.RandIntn(spread)
}

func makeAddrs() (string, string, string) {
	start := randPort()

	return fmt.Sprintf("tcp://0.0.0.0:%d", start),
		fmt.Sprintf("tcp://0.0.0.0:%d", start+1),
		fmt.Sprintf("tcp://0.0.0.0:%d", start+2)
}

// getConfig returns a config for test cases
func getConfig(t *testing.T) *cfg.Config {
	t.Helper()

	c := cfg.ResetTestRoot(t.Name())

	// and we use random ports to run in parallel
	tm, rpc, grpc := makeAddrs()
	c.P2P.ListenAddress = tm
	c.RPC.ListenAddress = rpc
	c.RPC.GRPCListenAddress = grpc

	return c
}

// heightStopWAL is a WAL which writes all msgs to underlying WALWriter.
// Writing stops when the heightToStop is reached. Client will be notified via
// signalWhenStopsTo channel.
type heightStopWAL struct {
	enc               *walm.WALWriter
	stopped           bool
	heightToStop      int64
	signalWhenStopsTo chan<- struct{}

	logger log.Logger
}

// needed for determinism
var fixedTime, _ = time.Parse(time.RFC3339, "2017-01-02T15:04:05Z")

func newHeightStopWAL(logger log.Logger, enc *walm.WALWriter, nBlocks int64, signalStop chan<- struct{}) *heightStopWAL {
	return &heightStopWAL{
		enc:               enc,
		heightToStop:      nBlocks,
		signalWhenStopsTo: signalStop,
		logger:            logger,
	}
}

func (w *heightStopWAL) SetLogger(logger log.Logger) {
	w.logger = logger
}

// Save writes message to the internal buffer except when heightToStop is
// reached, in which case it will signal the caller via signalWhenStopsTo and
// skip writing.
func (w *heightStopWAL) Write(m walm.WALMessage) error {
	if w.stopped {
		panic("WAL already stopped. Not writing meta message")
	}

	w.logger.Debug("WAL Write Message", "msg", m)
	err := w.enc.Write(walm.TimedWALMessage{fixedTime, m})
	if err != nil {
		panic(fmt.Sprintf("failed to encode the msg %v", m))
	}

	return nil
}

func (w *heightStopWAL) WriteSync(m walm.WALMessage) error {
	return w.Write(m)
}

func (w *heightStopWAL) WriteMetaSync(m walm.MetaMessage) error {
	if w.stopped {
		panic("WAL already stopped. Not writing meta message")
	}

	if m.Height != 0 {
		w.logger.Debug("WAL write end height message", "height", m.Height, "stopHeight", w.heightToStop)
		if m.Height == w.heightToStop {
			w.logger.Debug("Stopping WAL at height", "height", m.Height)
			w.signalWhenStopsTo <- struct{}{}
			w.stopped = true

			return nil
		}
	}

	// After processing is successful, commit to underlying store.  This must
	// come last.
	w.enc.WriteMeta(m)

	return nil
}

func (w *heightStopWAL) FlushAndSync() error { return nil }

func (w *heightStopWAL) SearchForHeight(height int64, options *walm.WALSearchOptions) (rd io.ReadCloser, found bool, err error) {
	return nil, false, nil
}

func (w *heightStopWAL) Start() error { return nil }
func (w *heightStopWAL) Stop() error  { return nil }
func (w *heightStopWAL) Wait()        {}
