package core

import (
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
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
	s += ufmt.Sprintf("\n## Client %s\n", c.id)
	s += ufmt.Sprintf("- Type: %s\n", c.typ)
	s += ufmt.Sprintf("- Creator: %s\n", c.creator)
	s += ufmt.Sprintf("- CounterpartyClientID: %s\n", c.counterpartyClientID)
	var merklePrefix string
	for _, m := range c.counterpartyMerklePrefix {
		merklePrefix += "'" + string(m) + "' "
	}
	s += ufmt.Sprintf("- CounterpartyMerklePrefix: %s\n", merklePrefix)
	return s
}
