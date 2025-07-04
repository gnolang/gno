package core

import (
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

type client struct {
	id          string
	lightClient lightclient.LightClientModule
}

func (cs *Store) addClient(typ string) *client {
	id := ufmt.Sprintf("%d-%d", typ, cs.clientSeq.Next())
	c := &client{id: id}
	switch typ {
	case lightclient.Tendermint:
		c.lightClient = &tendermint.TMLightClient{}
	default:
		panic("unhandled light client type " + typ)
	}
	cs.clientByID.Set(id, c)
	return c
}
