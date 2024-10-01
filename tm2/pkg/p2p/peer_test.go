package p2p

// func TestPeerBasic(t *testing.T) {
// 	t.Parallel()
//
// 	// simulate remote peer
// 	rp := &remotePeer{PrivKey: ed25519.GenPrivKey(), Config: p2p.cfg}
// 	rp.Start()
// 	defer rp.Stop()
//
// 	p, err := createOutboundPeerAndPerformHandshake(t, rp.Addr(), p2p.cfg, conn.DefaultMConnConfig())
// 	require.Nil(err)
//
// 	err = p.Start()
// 	require.Nil(err)
// 	defer p.Stop()
//
// 	assert.True(p.IsRunning())
// 	assert.True(p.IsOutbound())
// 	assert.False(p.IsPersistent())
// 	p.persistent = true
// 	assert.True(p.IsPersistent())
// 	assert.Equal(rp.Addr().DialString(), p.RemoteAddr().String())
// 	assert.Equal(rp.ID(), p.ID())
// }
//
// func TestPeerSend(t *testing.T) {
// 	t.Parallel()
//
// 	assert, require := assert.New(t), require.New(t)
//
// 	config := p2p.cfg
//
// 	// simulate remote peer
// 	rp := &remotePeer{PrivKey: ed25519.GenPrivKey(), Config: config}
// 	rp.Start()
// 	defer rp.Stop()
//
// 	p, err := createOutboundPeerAndPerformHandshake(t, rp.Addr(), config, conn.DefaultMConnConfig())
// 	require.Nil(err)
//
// 	err = p.Start()
// 	require.Nil(err)
//
// 	defer p.Stop()
//
// 	assert.True(p.Send(p2p.testCh, []byte("Asylum")))
// }
//
// func createOutboundPeerAndPerformHandshake(
// 	t *testing.T,
// 	addr *p2p.NetAddress,
// 	config *config.P2PConfig,
// 	mConfig conn.MConnConfig,
// ) (*peer, error) {
// 	t.Helper()
//
// 	chDescs := []*conn.ChannelDescriptor{
// 		{ID: p2p.testCh, Priority: 1},
// 	}
// 	reactorsByCh := map[byte]p2p.Reactor{p2p.testCh: p2p.NewTestReactor(chDescs, true)}
// 	pk := ed25519.GenPrivKey()
// 	pc, err := testOutboundPeerConn(addr, config, false, pk)
// 	if err != nil {
// 		return nil, err
// 	}
// 	timeout := 1 * time.Second
// 	ourNodeInfo := p2p.testNodeInfo(addr.ID, "host_peer")
// 	peerNodeInfo, err := p2p.handshake(pc.conn, timeout, ourNodeInfo)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	p := newPeer(pc, mConfig, peerNodeInfo, reactorsByCh, chDescs, func(p p2p.Peer, r interface{}) {})
// 	p.SetLogger(log.NewTestingLogger(t).With("peer", addr))
// 	return p, nil
// }
//
// func testDial(addr *p2p.NetAddress, cfg *config.P2PConfig) (net.Conn, error) {
// 	conn, err := addr.DialTimeout(cfg.DialTimeout)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conn, nil
// }
//
// func testOutboundPeerConn(
// 	addr *p2p.NetAddress,
// 	config *config.P2PConfig,
// 	persistent bool,
// 	ourNodePrivKey crypto.PrivKey,
// ) (peerConn, error) {
// 	var pc peerConn
// 	conn, err := testDial(addr, config)
// 	if err != nil {
// 		return pc, errors.Wrap(err, "Error creating peer")
// 	}
//
// 	pc, err = p2p.testPeerConn(conn, config, true, persistent, ourNodePrivKey, addr)
// 	if err != nil {
// 		if cerr := conn.Close(); cerr != nil {
// 			return pc, errors.Wrap(err, cerr.Error())
// 		}
// 		return pc, err
// 	}
//
// 	// ensure dialed ID matches connection ID
// 	if addr.ID != pc.ID() {
// 		if cerr := conn.Close(); cerr != nil {
// 			return pc, errors.Wrap(err, cerr.Error())
// 		}
// 		return pc, p2p.SwitchAuthenticationFailureError{addr, pc.ID()}
// 	}
//
// 	return pc, nil
// }
//
// type remotePeer struct {
// 	PrivKey    crypto.PrivKey
// 	Config     *config.P2PConfig
// 	addr       *p2p.NetAddress
// 	channels   []byte
// 	listenAddr string
// 	listener   net.Listener
// }
//
// func (rp *remotePeer) Addr() *p2p.NetAddress {
// 	return rp.addr
// }
//
// func (rp *remotePeer) ID() p2p.ID {
// 	return rp.PrivKey.PubKey().Address().ID()
// }
//
// func (rp *remotePeer) Start() error {
// 	if rp.listenAddr == "" {
// 		rp.listenAddr = "127.0.0.1:0"
// 	}
//
// 	l, err := net.Listen("tcp", rp.listenAddr) // any available address
// 	if err != nil {
// 		golog.Fatalf("net.Listen tcp :0: %+v", err)
//
// 		return err
// 	}
//
// 	rp.listener = l
// 	rp.addr, err = p2p.NewNetAddress(rp.PrivKey.PubKey().Address().ID(), l.Addr())
// 	if err != nil {
// 		return err
// 	}
//
// 	if rp.channels == nil {
// 		rp.channels = []byte{p2p.testCh}
// 	}
// 	go rp.accept()
//
// 	return nil
// }
//
// func (rp *remotePeer) Stop() {
// 	rp.listener.Close()
// }
//
// func (rp *remotePeer) Dial(addr *p2p.NetAddress) (net.Conn, error) {
// 	conn, err := addr.DialTimeout(1 * time.Second)
// 	if err != nil {
// 		return nil, err
// 	}
// 	pc, err := p2p.testInboundPeerConn(conn, rp.Config, rp.PrivKey)
// 	if err != nil {
// 		return nil, err
// 	}
// 	_, err = p2p.handshake(pc.conn, time.Second, rp.nodeInfo())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conn, err
// }
//
// func (rp *remotePeer) accept() {
// 	conns := []net.Conn{}
//
// 	for {
// 		conn, err := rp.listener.Accept()
// 		if err != nil {
// 			golog.Printf("Failed to accept conn: %+v", err)
// 			for _, conn := range conns {
// 				_ = conn.Close()
// 			}
// 			return
// 		}
//
// 		pc, err := p2p.testInboundPeerConn(conn, rp.Config, rp.PrivKey)
// 		if err != nil {
// 			golog.Fatalf("Failed to create a peer: %+v", err)
// 		}
//
// 		_, err = p2p.handshake(pc.conn, time.Second, rp.nodeInfo())
// 		if err != nil {
// 			golog.Fatalf("Failed to perform handshake: %+v", err)
// 		}
//
// 		conns = append(conns, conn)
// 	}
// }
//
// func (rp *remotePeer) nodeInfo() p2p.NodeInfo {
// 	return p2p.NodeInfo{
// 		VersionSet: p2p.testVersionSet(),
// 		NetAddress: rp.Addr(),
// 		Network:    "testing",
// 		Version:    "1.2.3-rc0-deadbeef",
// 		Channels:   rp.channels,
// 		Moniker:    "remote_peer",
// 	}
// }
