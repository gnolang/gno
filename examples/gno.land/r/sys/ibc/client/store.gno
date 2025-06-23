package client

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/sys/ibc/lightclient"
	"gno.land/p/sys/ibc/lightclient/tendermint"
)

type clientStore struct {
	seq  seqid.ID
	byID *avl.Tree
}

func newClientStore() *clientStore {
	return &clientStore{
		byID: avl.NewTree(),
	}
}

type client struct {
	id          string
	lightClient lightclient.LightClientModule
}

func (cs *clientStore) add(typ string) client {
	id := ufmt.Sprintf("%d-%d", typ, cs.seq.Next())
	c := &client{id: id}
	switch typ {
	case Tendermint:
		c.lightClient = &tendermint.TMLightClient{}
	default:
		panic("unhandled light client type " + typ)
	}
	cs.byID.Set(id, c)
	return c
}
