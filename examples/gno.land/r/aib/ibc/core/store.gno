package core

import (
	"chain/params"

	"gno.land/p/aib/ibc/host"
	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
)

var store *Store

func init() {
	store = newStore()
}

type Store struct {
	clientSeq  seqid.ID
	clientByID *avl.Tree // id:client
}

func newStore() *Store {
	return &Store{
		clientByID: avl.NewTree(),
	}
}

type client struct {
	id                       string
	typ                      string
	creator                  address
	counterpartyClientID     string
	counterpartyMerklePrefix [][]byte

	lightClient lightclient.Interface

	sendSeq                seqid.ID
	packetCommitmentsBySeq *avl.Tree // sequence:commitment
}

// setPacketCommitment stores the packet commitment in the client data and in
// the params.
// CONTRACT: c.sendSeq has been incremented and contains the current sequence
// number.
func (c *client) setPacketCommitment(packet types.Packet) {
	commitment := types.CommitPacket(packet)

	// store in client data
	c.packetCommitmentsBySeq.Set(c.sendSeq.Binary(), commitment)

	// store in params, so it is provable by the counterparty
	// The param key will be:
	// "vm:gno.land/r/aib/ibc/core:" + clientID + "1" + sequence
	key := host.PacketCommitmentKey(c.id, uint64(c.sendSeq))
	// TODO make host.PacketCommitmentKey returns string ?
	// TODO find a way to assert this write in tests
	params.SetBytes(string(key), commitment)
}

func (cs *Store) addClient(typ string, creator address) *client {
	id := ufmt.Sprintf("%s-%d", typ, uint64(cs.clientSeq.Next()))
	c := &client{
		id:      id,
		typ:     typ,
		creator: creator,

		packetCommitmentsBySeq: avl.NewTree(),
	}
	switch typ {
	case lightclient.Tendermint:
		c.lightClient = &tendermint.TMLightClient{}
	default:
		panic("unhandled light client type " + typ)
	}
	cs.clientByID.Set(id, c)
	return c
}

func (cs *Store) getClient(id string) *client {
	x, found := cs.clientByID.Get(id)
	if !found {
		return nil
	}
	return x.(*client)
}
