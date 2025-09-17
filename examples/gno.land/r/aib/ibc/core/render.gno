package core

import (
	"encoding/hex"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
)

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", RenderHome)
	router.HandleFunc("clients", RenderClients)
	router.HandleFunc("clients/{id}", RenderClient)
	return router.Render(path)
}

func RenderHome(w *mux.ResponseWriter, r *mux.Request) {
	w.Write("Hello IBC!")
}

func RenderClients(w *mux.ResponseWriter, r *mux.Request) {
	store.clientByID.Iterate("", "", func(key string, v any) bool {
		c := v.(*client)
		w.Write(c.Render())
		return false
	})
}

func RenderClient(w *mux.ResponseWriter, r *mux.Request) {
	id := r.GetVar("id")
	c, found := store.clientByID.Get(id)
	if !found {
		w.Write(ufmt.Sprintf("client %s not found", id))
		return
	}
	w.Write(c.(*client).Render())
}

func (c *client) Render() string {
	var s string
	s += ufmt.Sprintf("## Client %s\n", c.id)
	s += ufmt.Sprintf("- Type: %s\n", c.typ)
	s += ufmt.Sprintf("- Creator: %s\n", c.creator)
	s += ufmt.Sprintf("- CounterpartyClientID: %s\n", c.counterpartyClientID)
	var merklePrefix string
	for _, m := range c.counterpartyMerklePrefix {
		merklePrefix += "'" + string(m) + "' "
	}
	s += ufmt.Sprintf("- CounterpartyMerklePrefix: %s\n", merklePrefix)
	switch c.typ {
	case lightclient.Tendermint:
		lc := c.lightClient.(*tendermint.TMLightClient)
		latestHeight := lc.LatestHeight().String()
		x, found := lc.ConsensusStateByHeight.Get(latestHeight)
		if !found {
			s += "Latest consensus state not found\n"
			return s
		}
		latestConsState := x.(*tendermint.ConsensusState)
		s += "- Latest consensus state:\n"
		s += ufmt.Sprintf("  - Height: %s\n", latestHeight)
		s += ufmt.Sprintf("  - Timestamp: %s\n", latestConsState.Timestamp)
		s += ufmt.Sprintf("  - Root: %s\n", hex.EncodeToString(latestConsState.Root.Hash))
		s += ufmt.Sprintf("  - NextValidatorsHash: %s\n", hex.EncodeToString(latestConsState.NextValidatorsHash))
	}
	return s
}
