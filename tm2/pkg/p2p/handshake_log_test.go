package p2p

import (
	"context"
	"encoding/binary"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/require"
)

// maxHandshakeLog bounds what one failed handshake may put in the log.
const maxHandshakeLog = 1024

// nodeInfoWithBadPort returns a sized NodeInfo holding only an address of 10000
// control bytes in its port, which fails to decode with an error quoting them.
func nodeInfoWithBadPort(id types.ID) []byte {
	addr := []byte(string(id) + "@1.2.3.4:" + strings.Repeat("\x01", 10_000))
	body := append(binary.AppendUvarint([]byte{2<<3 | 2}, uint64(len(addr))), addr...)

	return append(binary.AppendUvarint(nil, uint64(len(body))), body...)
}

func TestHandshakeBoundsLogOfPeerNodeInfo(t *testing.T) {
	t.Parallel()

	t.Run("accepting", func(t *testing.T) {
		t.Parallel()

		nodeKey := types.GenerateNodeKey()
		addr, err := types.NewNetAddress(nodeKey.ID(), &net.TCPAddr{IP: net.ParseIP("127.0.0.1")})
		require.NoError(t, err)

		logs := &lockedBuffer{}
		tr := NewMultiplexTransport(
			types.NodeInfo{Network: "testing", NetAddress: addr},
			*nodeKey,
			conn.DefaultMConnConfig(),
			slog.New(slog.NewTextHandler(logs, nil)),
		)
		require.NoError(t, tr.Listen(*addr))
		t.Cleanup(func() { _ = tr.Close() })

		listening := tr.NetAddress()
		c, err := net.Dial("tcp", listening.DialString())
		require.NoError(t, err)
		defer c.Close()

		key := types.GenerateNodeKey()
		sc, err := conn.MakeSecretConnection(c, key.PrivKey)
		require.NoError(t, err)

		_, err = sc.Write(nodeInfoWithBadPort(key.ID()))
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			return strings.Contains(logs.String(), "unable to process p2p connection")
		}, 10*time.Second, 10*time.Millisecond)
		require.Contains(t, logs.String(), "unable to parse port")
		require.LessOrEqual(t, len(logs.String()), maxHandshakeLog)
	})

	t.Run("dialing", func(t *testing.T) {
		t.Parallel()

		sw := NewMultiplexSwitch(newLoopbackTransport(t, "testing", "testing"))
		logs := captureLogs(sw)

		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer l.Close()

		key := types.GenerateNodeKey()

		done := make(chan struct{})
		defer close(done)

		go func() {
			c, err := l.Accept()
			if err != nil {
				return
			}
			defer c.Close()

			sc, err := conn.MakeSecretConnection(c, key.PrivKey)
			if err != nil {
				return
			}

			_, _ = sc.Write(nodeInfoWithBadPort(key.ID()))

			// The connection stays open until the dialer has read the message.
			<-done
		}()

		addr, err := types.NewNetAddress(key.ID(), l.Addr())
		require.NoError(t, err)

		sw.dialPeer(context.Background(), addr)

		require.Contains(t, logs.String(), "unable to dial peer")
		require.Contains(t, logs.String(), "unable to parse port")
		require.LessOrEqual(t, len(logs.String()), maxHandshakeLog)
	})
}
