package core

import (
	"std"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

var store *Store

func init() {
	store = newStore()
}

type Store struct {
	clientSeq  seqid.ID
	clientByID *avl.Tree
}

func newStore() *Store {
	return &Store{
		clientByID: avl.NewTree(),
	}
}

func (cs *Store) addClient(typ string, creator std.Address) *client {
	id := ufmt.Sprintf("%s-%d", typ, uint64(cs.clientSeq.Next()))
	c := &client{
		id:      id,
		typ:     typ,
		creator: creator,
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
