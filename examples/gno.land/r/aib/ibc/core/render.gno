package core

import (
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
)

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", RenderHome)
	return router.Render(path)
}

func RenderHome(w *mux.ResponseWriter, r *mux.Request) {
	store.clientByID.Iterate("", "", func(key string, v any) bool {
		c := v.(*client)
		w.Write(c.Render())
		return false
	})
}

func (c *client) Render() string {
	var s string
	s += ufmt.Sprintf("## Client %s\n", c.id)
	s += ufmt.Sprintf("- Type: %s\n", c.typ)
	s += ufmt.Sprintf("- Creator: %s\n", c.creator)
	s += ufmt.Sprintf("- CounterpartyClientID: %s\n", c.counterpartyClientID)
	var merklePrefix string
	for m := range c.counterpartyMerklePrefix {
		merklePrefix += ufmt.Sprint(m)
		merklePrefix += " "
	}
	s += ufmt.Sprintf("- CounterpartyMerklePrefix: %s\n", merklePrefix)
	return s
}
