package mempool

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	memcfg "github.com/gnolang/gno/tm2/pkg/bft/mempool/config"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/stretchr/testify/require"
)

// TestReceiveBoundsLogOfPeerMessage sends a 1 MB transaction, which the
// mempool admits, and messages that fail decoding, and bounds what each puts
// in the log at debug level, the app's callback for the transaction included.
func TestReceiveBoundsLogOfPeerMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		msg  []byte
		stop bool
	}{
		{"transaction", amino.MustMarshalAny(&TxMessage{Tx: make([]byte, 1_000_000)}), false},
		{"non-ASCII type URL", p2pTesting.BadTypeURLMessage(1_000_000), true},
		{"malformed message of another module", p2pTesting.MalformedMessage(1_000_000), true},
	}

	appConn, err := proxy.NewLocalClientCreator(kvstore.NewKVStoreApplication()).NewABCIClient()
	require.NoError(t, err)
	require.NoError(t, appConn.Start())
	defer appConn.Stop()
	mempool := NewCListMempool(memcfg.TestMempoolConfig(), appConn, 0, types.MaxBlockTxBytes)

	var logged bytes.Buffer
	sw := &p2pTesting.StopRecorder{}

	reactor := NewReactor(memcfg.TestMempoolConfig(), mempool)
	reactor.SetLogger(p2pTesting.DebugLogger(&logged))
	reactor.SetSwitch(sw)

	peer := p2pTesting.NewPeer(t)
	defer peer.Stop()
	reactor.InitPeer(peer)

	for _, tc := range cases {
		p2pTesting.RequireBoundedReceive(t, &logged, sw, tc.name, tc.stop, func() {
			reactor.Receive(MempoolChannel, peer, tc.msg)
		})
	}
	require.Equal(t, 1, mempool.Size(), "the transaction did not reach the app's callback")
}
