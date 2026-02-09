package consensus

import (
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	walm "github.com/gnolang/gno/tm2/pkg/bft/wal"
	"github.com/gnolang/gno/tm2/pkg/random"
)

func randPort() int {
	// returns between base and base + spread
	base, spread := 20000, 20000
	return base + random.RandIntn(spread)
}

func makeAddrs() (string, string) {
	start := randPort()
	return fmt.Sprintf("0.0.0.0:%d", start),
		fmt.Sprintf("0.0.0.0:%d", start+1)
}

// getConfig returns a config and genesis file for test cases
func getConfig(t *testing.T) (*cfg.Config, string) {
	t.Helper()

	c, genesisFile := cfg.ResetTestRoot(t.Name())

	// and we use random ports to run in parallel
	tm, rpc := makeAddrs()
	c.P2P.ListenAddress = tm
	c.RPC.ListenAddress = rpc

	return c, genesisFile
}

// heightStopWAL is a WAL which writes all msgs to underlying WALWriter.
// Writing stops when the heightToStop is reached. Client will be notified via
// signalWhenStopsTo channel.
type heightStopWAL struct {
	enc               *walm.WALWriter
	stopped           bool
	heightToStop      int64
	signalWhenStopsTo chan<- struct{}

	logger *slog.Logger
}

// needed for determinism
var fixedTime, _ = time.Parse(time.RFC3339, "2017-01-02T15:04:05Z")

func newHeightStopWAL(logger *slog.Logger, enc *walm.WALWriter, nBlocks int64, signalStop chan<- struct{}) *heightStopWAL {
	return &heightStopWAL{
		enc:               enc,
		heightToStop:      nBlocks,
		signalWhenStopsTo: signalStop,
		logger:            logger,
	}
}

func (w *heightStopWAL) SetLogger(logger *slog.Logger) {
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
	err := w.enc.Write(walm.TimedWALMessage{Time: fixedTime, Msg: m})
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
