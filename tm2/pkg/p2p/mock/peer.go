package mock

import (
	"net"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/service"
)

type Peer struct {
	*service.BaseService
	ip                   net.IP
	id                   p2p.ID
	addr                 *types.NetAddress
	kv                   map[string]interface{}
	Outbound, Persistent bool
}

// NewPeer creates and starts a new mock peer. If the ip
// is nil, random routable address is used.
func NewPeer(ip net.IP) *Peer {
	var netAddr *types.NetAddress
	if ip == nil {
		_, netAddr = p2p.CreateRoutableAddr()
	} else {
		netAddr = types.NewNetAddressFromIPPort(ip, 26656)
	}
	nodeKey := types.NodeKey{PrivKey: ed25519.GenPrivKey()}
	netAddr.ID = nodeKey.ID()
	mp := &Peer{
		ip:   ip,
		id:   nodeKey.ID(),
		addr: netAddr,
		kv:   make(map[string]interface{}),
	}
	mp.BaseService = service.NewBaseService(nil, "MockPeer", mp)
	mp.Start()
	return mp
}

func (mp *Peer) FlushStop()                              { mp.Stop() }
func (mp *Peer) TrySend(chID byte, msgBytes []byte) bool { return true }
func (mp *Peer) Send(chID byte, msgBytes []byte) bool    { return true }
func (mp *Peer) NodeInfo() types.NodeInfo {
	return types.NodeInfo{
		NetAddress: mp.addr,
	}
}
func (mp *Peer) Status() multiplex.ConnectionStatus { return multiplex.ConnectionStatus{} }
func (mp *Peer) ID() p2p.ID                         { return mp.id }
func (mp *Peer) IsOutbound() bool                   { return mp.Outbound }
func (mp *Peer) IsPersistent() bool                 { return mp.Persistent }
func (mp *Peer) Get(key string) interface{} {
	if value, ok := mp.kv[key]; ok {
		return value
	}
	return nil
}

func (mp *Peer) Set(key string, value interface{}) {
	mp.kv[key] = value
}
func (mp *Peer) RemoteIP() net.IP              { return mp.ip }
func (mp *Peer) SocketAddr() *types.NetAddress { return mp.addr }
func (mp *Peer) RemoteAddr() net.Addr          { return &net.TCPAddr{IP: mp.ip, Port: 8800} }
func (mp *Peer) CloseConn() error              { return nil }
