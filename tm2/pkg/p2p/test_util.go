package p2p

import (
	"fmt"
	"net"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p/config"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/gnolang/gno/tm2/pkg/versionset"
)

const testCh = 0x01

// ------------------------------------------------

func CreateRoutableAddr() (addr string, netAddr *NetAddress) {
	for {
		id := ed25519.GenPrivKey().PubKey().Address().ID()
		var err error
		addr = fmt.Sprintf("%s@%v.%v.%v.%v:26656", id, random.RandInt()%256, random.RandInt()%256, random.RandInt()%256, random.RandInt()%256)
		netAddr, err = NewNetAddressFromString(addr)
		if err != nil {
			panic(err)
		}
		if netAddr.Routable() {
			break
		}
	}
	return
}

// ------------------------------------------------------------------
// Connects switches via arbitrary net.Conn. Used for testing.

const TEST_HOST = "localhost"

// MakeConnectedSwitches returns n switches, connected according to the connect func.
// If connect==Connect2Switches, the switches will be fully connected.
// initSwitch defines how the i'th switch should be initialized (ie. with what reactors).
// NOTE: panics if any switch fails to start.
func MakeConnectedSwitches(cfg *config.P2PConfig, n int, initSwitch func(int, *Switch) *Switch, connect func([]*Switch, int, int)) []*Switch {
	switches := make([]*Switch, n)
	for i := 0; i < n; i++ {
		switches[i] = MakeSwitch(cfg, i, TEST_HOST, "123.123.123", initSwitch)
	}

	if err := StartSwitches(switches); err != nil {
		panic(err)
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			connect(switches, i, j)
		}
	}

	return switches
}

// Connect2Switches will connect switches i and j via net.Pipe().
// Blocks until a connection is established.
// NOTE: caller ensures i and j are within bounds.
func Connect2Switches(switches []*Switch, i, j int) {
	switchI := switches[i]
	switchJ := switches[j]

	c1, c2 := conn.NetPipe()

	doneCh := make(chan struct{})
	go func() {
		err := switchI.addPeerWithConnection(c1)
		if err != nil {
			panic(err)
		}
		doneCh <- struct{}{}
	}()
	go func() {
		err := switchJ.addPeerWithConnection(c2)
		if err != nil {
			panic(err)
		}
		doneCh <- struct{}{}
	}()
	<-doneCh
	<-doneCh
}

func (sw *Switch) addPeerWithConnection(conn net.Conn) error {
	pc, err := testInboundPeerConn(conn, sw.config, sw.nodeKey.PrivKey)
	if err != nil {
		if err := conn.Close(); err != nil {
			sw.Logger.Error("Error closing connection", "err", err)
		}
		return err
	}

	ni, err := handshake(conn, time.Second, sw.nodeInfo)
	if err != nil {
		if err := conn.Close(); err != nil {
			sw.Logger.Error("Error closing connection", "err", err)
		}
		return err
	}

	p := newPeer(
		pc,
		MConnConfig(sw.config),
		ni,
		sw.reactorsByCh,
		sw.chDescs,
		sw.StopPeerForError,
	)

	if err = sw.addPeer(p); err != nil {
		pc.CloseConn()
		return err
	}

	return nil
}

// StartSwitches calls sw.Start() for each given switch.
// It returns the first encountered error.
func StartSwitches(switches []*Switch) error {
	for _, s := range switches {
		err := s.Start() // start switch and reactors
		if err != nil {
			return err
		}
	}
	return nil
}

func MakeSwitch(
	cfg *config.P2PConfig,
	i int,
	network, version string,
	initSwitch func(int, *Switch) *Switch,
	opts ...SwitchOption,
) *Switch {
	nodeKey := NodeKey{
		PrivKey: ed25519.GenPrivKey(),
	}
	nodeInfo := testNodeInfo(nodeKey.ID(), fmt.Sprintf("node%d", i))

	t := NewMultiplexTransport(nodeInfo, nodeKey, MConnConfig(cfg))

	if err := t.Listen(*nodeInfo.NetAddress); err != nil {
		panic(err)
	}

	// TODO: let the config be passed in?
	sw := initSwitch(i, NewSwitch(cfg, t, opts...))
	sw.SetLogger(log.NewNoopLogger().With("switch", i))
	sw.SetNodeKey(&nodeKey)

	for ch := range sw.reactorsByCh {
		nodeInfo.Channels = append(nodeInfo.Channels, ch)
	}

	// TODO: We need to setup reactors ahead of time so the NodeInfo is properly
	// populated and we don't have to do those awkward overrides and setters.
	t.nodeInfo = nodeInfo
	sw.SetNodeInfo(nodeInfo)

	return sw
}

func testInboundPeerConn(
	conn net.Conn,
	config *config.P2PConfig,
	ourNodePrivKey crypto.PrivKey,
) (peerConn, error) {
	return testPeerConn(conn, config, false, false, ourNodePrivKey, nil)
}

func testPeerConn(
	rawConn net.Conn,
	cfg *config.P2PConfig,
	outbound, persistent bool,
	ourNodePrivKey crypto.PrivKey,
	socketAddr *NetAddress,
) (pc peerConn, err error) {
	conn := rawConn

	// Fuzz connection
	if cfg.TestFuzz {
		// so we have time to do peer handshakes and get set up
		conn = FuzzConnAfterFromConfig(conn, 10*time.Second, cfg.TestFuzzConfig)
	}

	// Encrypt connection
	conn, err = upgradeSecretConn(conn, cfg.HandshakeTimeout, ourNodePrivKey)
	if err != nil {
		return pc, errors.Wrap(err, "Error creating peer")
	}

	// Only the information we already have
	return newPeerConn(outbound, persistent, conn, socketAddr), nil
}

// ----------------------------------------------------------------
// rand node info

func testNodeInfo(id ID, name string) NodeInfo {
	return testNodeInfoWithNetwork(id, name, "testing")
}

func testVersionSet() versionset.VersionSet {
	return versionset.VersionSet{
		versionset.VersionInfo{
			Name:    "p2p",
			Version: "v0.0.0", // dontcare
		},
	}
}

func testNodeInfoWithNetwork(id ID, name, network string) NodeInfo {
	return NodeInfo{
		VersionSet: testVersionSet(),
		NetAddress: NewNetAddressFromIPPort(id, net.ParseIP("127.0.0.1"), 0),
		Network:    network,
		Software:   "p2ptest",
		Version:    "v1.2.3-rc.0-deadbeef",
		Channels:   []byte{testCh},
		Moniker:    name,
		Other: NodeInfoOther{
			TxIndex:    "on",
			RPCAddress: fmt.Sprintf("127.0.0.1:%d", 0),
		},
	}
}
