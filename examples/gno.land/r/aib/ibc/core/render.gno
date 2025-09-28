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
	router.HandleFunc("clients/{id}/consensus_states", RenderClientConsensusStates)
	return router.Render(path)
}

func RenderHome(w *mux.ResponseWriter, r *mux.Request) {
	w.Write("Hello IBC!")
}

func RenderClients(w *mux.ResponseWriter, r *mux.Request) {
	store.clientByID.Iterate("", "", func(_ string, v any) bool {
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

func RenderClientConsensusStates(w *mux.ResponseWriter, r *mux.Request) {
	id := r.GetVar("id")
	c, found := store.clientByID.Get(id)
	if !found {
		w.Write(ufmt.Sprintf("client %s not found", id))
		return
	}
	w.Write(c.(*client).RenderConsensusStates())
}

func (c *client) Render() string {
	var s string
	s += ufmt.Sprintf("## Client %s\n", c.id)
	s += ufmt.Sprintf("- Type: %s\n", c.typ)
	s += ufmt.Sprintf("- Creator: %s\n", c.creator)
	s += ufmt.Sprintf("- Status: %s\n", c.lightClient.Status())
	s += ufmt.Sprintf("- CounterpartyClientID: %s\n", c.counterpartyClientID)
	var merklePrefix string
	for _, m := range c.counterpartyMerklePrefix {
		merklePrefix += "'" + string(m) + "' "
	}
	s += ufmt.Sprintf("- CounterpartyMerklePrefix: %s\n", merklePrefix)
	switch c.typ {
	case lightclient.Tendermint:
		lc := c.lightClient.(*tendermint.TMLightClient)
		s += "- Client state:\n"
		s += ufmt.Sprintf("  - ChainId: %s\n", lc.ClientState.ChainID)
		s += ufmt.Sprintf("  - TrustingPeriod: %s\n", lc.ClientState.TrustingPeriod)
		s += ufmt.Sprintf("  - UnbondingPeriod: %s\n", lc.ClientState.UnbondingPeriod)
		s += ufmt.Sprintf("  - LatestHeight: %s\n", lc.LatestHeight().String())
		latestConsState, found := lc.GetConsensusState(lc.LatestHeight())
		if !found {
			s += "Latest consensus state not found\n"
			return s
		}
		s += "- Latest consensus state:\n"
		s += ufmt.Sprintf("  - Timestamp: %s\n", latestConsState.Timestamp)
		s += ufmt.Sprintf("  - Root: %s\n", hex.EncodeToString(latestConsState.Root.Hash))
		s += ufmt.Sprintf("  - NextValidatorsHash: %s\n", hex.EncodeToString(latestConsState.NextValidatorsHash))
	}
	return s
}

func (c *client) RenderConsensusStates() string {
	var s string
	s += ufmt.Sprintf("## Client %s consensus states\n", c.id)
	switch c.typ {
	case lightclient.Tendermint:
		lc := c.lightClient.(*tendermint.TMLightClient)
		lc.ConsensusStateByHeight.Iterate("", "", func(key string, v any) bool {
			s += ufmt.Sprintf("- ConsensusState height %s:\n", key)
			cs := v.(*tendermint.ConsensusState)
			s += ufmt.Sprintf("  - Timestamp: %s\n", cs.Timestamp)
			s += ufmt.Sprintf("  - RootHash: %s\n", hex.EncodeToString(cs.Root.Hash))
			s += ufmt.Sprintf("  - NextValidatorsHash: %s\n", hex.EncodeToString(cs.NextValidatorsHash))
			return false
		})
	}
	return s
}
