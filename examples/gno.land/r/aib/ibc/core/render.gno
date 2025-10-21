package core

import (
	"encoding/base64"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
	"gno.land/p/onbloc/json"
)

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", RenderHome)
	router.HandleFunc("apps", RenderApps)
	router.HandleFunc("clients", RenderClients)
	router.HandleFunc("clients/{id}", RenderClient)
	router.HandleFunc("clients/{id}/consensus_states", RenderClientConsensusStates)
	router.HandleFunc("clients/{id}/consensus_states/{revision_number}/{revision_height}", RenderClientConsensusState)
	router.HandleFunc("clients/{id}/next_sequence_send", RenderClientNextSequenceSend)
	router.HandleFunc("clients/{id}/packet_commitments", RenderClientPacketCommitments)
	return router.Render(path)
}

func RenderHome(w *mux.ResponseWriter, r *mux.Request) {
	// TODO list available endpoints ?
	w.Write(`["Hello IBC!"]`)
}

func RenderApps(w *mux.ResponseWriter, r *mux.Request) {
	var nodes []*json.Node
	for port, app := range store.routes {
		nodes = append(nodes, json.ObjectNode("", map[string]*json.Node{
			"port_id":  json.StringNode("", port),
			"pkg_path": json.StringNode("", app.pkgPath),
			"address":  json.StringNode("", app.address.String()),
		}))
	}
	renderNode(w, json.ArrayNode("", nodes))
}

func RenderClients(w *mux.ResponseWriter, r *mux.Request) {
	var nodes []*json.Node
	store.clientByID.Iterate("", "", func(_ string, v any) bool {
		// TODO paginate
		c := v.(*client)
		nodes = append(nodes, c.Render())
		return false
	})
	renderNode(w, json.ArrayNode("", nodes))
}

func RenderClient(w *mux.ResponseWriter, r *mux.Request) {
	id := r.GetVar("id")
	c := store.getClient(id)
	if c == nil {
		renderNode(w, nodeErrorClientNotFound(id))
		return
	}
	renderNode(w, c.Render())
}

func RenderClientConsensusStates(w *mux.ResponseWriter, r *mux.Request) {
	id := r.GetVar("id")
	c := store.getClient(id)
	if c == nil {
		renderNode(w, nodeErrorClientNotFound(id))
		return
	}
	renderNode(w, c.RenderConsensusStates())
}

func RenderClientConsensusState(w *mux.ResponseWriter, r *mux.Request) {
	var (
		id             = r.GetVar("id")
		revisionNumber = r.GetVar("revision_number")
		revisionHeight = r.GetVar("revision_height")
		heightStr      = revisionNumber + "/" + revisionHeight
	)
	height, err := types.ParseHeight(heightStr)
	if err != nil {
		renderNode(w, nodeError("cant parse height %s: %v", heightStr, err))
		return
	}
	c := store.getClient(id)
	if c == nil {
		renderNode(w, nodeErrorClientNotFound(id))
		return
	}
	switch c.typ {
	case lightclient.Tendermint:
		lc := c.lightClient.(*tendermint.TMLightClient)
		cs, found := lc.GetConsensusState(height)
		if !found {
			renderNode(w, nodeError("consensus state not found for height %s", height))
			return
		}
		renderNode(w, renderConsensusState(height, cs))
	}
}

func RenderClientNextSequenceSend(w *mux.ResponseWriter, r *mux.Request) {
	id := r.GetVar("id")
	c := store.getClient(id)
	if c == nil {
		renderNode(w, nodeErrorClientNotFound(id))
		return
	}
	w.Write(ufmt.Sprintf(`{"next_sequence_send": %d}"`, int64(c.sendSeq)+1))
}

func RenderClientPacketCommitments(w *mux.ResponseWriter, r *mux.Request) {
	id := r.GetVar("id")
	c := store.getClient(id)
	if c == nil {
		renderNode(w, nodeErrorClientNotFound(id))
		return
	}
	var nodes []*json.Node
	c.packetCommitmentsBySeq.Iterate("", "", func(key string, v any) bool {
		// TODO paginate
		id, _ := seqid.FromBinary(key)
		nodes = append(nodes, json.ObjectNode("", map[string]*json.Node{
			"sequence": json.StringNode("", id.String()),
			"data":     json.StringNode("", base64.StdEncoding.EncodeToString(v.([]byte))),
		}))
		return false
	})
	renderNode(w, json.ArrayNode("", nodes))
}

func (c *client) Render() *json.Node {
	m := map[string]*json.Node{
		"id":                     json.StringNode("", c.id),
		"type":                   json.StringNode("", c.typ),
		"creator":                json.StringNode("", c.creator.String()),
		"status":                 json.StringNode("", c.lightClient.Status()),
		"counterparty_client_id": json.StringNode("", c.counterpartyClientID),
	}
	var prefixes []*json.Node
	for _, m := range c.counterpartyMerklePrefix {
		prefixes = append(prefixes, json.StringNode("", string(m)))
	}
	m["counterparty_merke_prefix"] = json.ArrayNode("", prefixes)
	switch c.typ {
	case lightclient.Tendermint:
		lc := c.lightClient.(*tendermint.TMLightClient)
		m["client_state"] = json.ObjectNode("", map[string]*json.Node{
			"chain_id":         json.StringNode("", lc.ClientState.ChainID),
			"latest_height":    renderHeight(lc.ClientState.LatestHeight),
			"frozen_height":    renderHeight(lc.ClientState.FrozenHeight),
			"trust_level":      renderFraction(lc.ClientState.TrustLevel),
			"trusting_period":  json.StringNode("", lc.ClientState.TrustingPeriod.String()),
			"unbonding_period": json.StringNode("", lc.ClientState.UnbondingPeriod.String()),
			"max_clock_drift":  json.StringNode("", lc.ClientState.MaxClockDrift.String()),
			"upgrade_path":     renderStrings(lc.ClientState.UpgradePath),
		})
		lastConsState, found := lc.GetConsensusState(lc.LatestHeight())
		if found {
			m["last_consensus_state"] = renderConsensusState(lc.LatestHeight(), lastConsState)
		}
	}

	return json.ObjectNode("", m)
}

func (c *client) RenderConsensusStates() *json.Node {
	switch c.typ {
	case lightclient.Tendermint:
		lc := c.lightClient.(*tendermint.TMLightClient)
		var nodes []*json.Node
		lc.ConsensusStateByHeight.Iterate("", "", func(key string, v any) bool {
			cs := v.(*tendermint.ConsensusState)
			height, _ := types.ParseHeight(key)
			nodes = append(nodes, renderConsensusState(height, cs))
			return false
		})
		return json.ArrayNode("", nodes)
	}
	return nodeError("unhandled client type %s", c.typ)
}

func renderConsensusState(height types.Height, cs *tendermint.ConsensusState) *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"height":               renderHeight(height),
		"timestamp":            json.NumberNode("", float64(cs.Timestamp.Unix())),
		"root":                 json.StringNode("", base64.StdEncoding.EncodeToString(cs.Root.Hash)),
		"next_validators_hash": json.StringNode("", base64.StdEncoding.EncodeToString(cs.NextValidatorsHash)),
	})
}

func renderHeight(h types.Height) *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"revision_number": json.NumberNode("", float64(h.RevisionNumber)),
		"revision_height": json.NumberNode("", float64(h.RevisionHeight)),
	})
}

func renderFraction(f tendermint.Fraction) *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"numerator":   json.NumberNode("", float64(f.Numerator)),
		"denominator": json.NumberNode("", float64(f.Denominator)),
	})
}

func renderStrings(s []string) *json.Node {
	var nodes []*json.Node
	for _, s := range s {
		nodes = append(nodes, json.StringNode("", s))
	}
	return json.ArrayNode("", nodes)
}

func renderNode(w *mux.ResponseWriter, n *json.Node) {
	bz, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	w.Write(string(bz))
}

func nodeErrorClientNotFound(id string) *json.Node {
	return nodeError("client %s not found", id)
}

func nodeError(msg string, args ...any) *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"error": json.StringNode("", ufmt.Sprintf(msg, args...)),
	})
}
