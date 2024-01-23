package agent

import "gno.land/p/demo/avl"

type Whitelist struct {
	store *avl.Tree
}

func (m *Whitelist) ClearAddresses() {
	m.store = nil
}

func (m *Whitelist) AddAddresses(addresses []string) {
	if m.store == nil {
		m.store = avl.NewTree()
	}

	for _, address := range addresses {
		m.store.Set(address, struct{}{})
	}
}

func (m *Whitelist) RemoveAddress(address string) {
	if m.store == nil {
		return
	}

	m.store.Remove(address)
}

func (m Whitelist) HasDefinition() bool {
	return m.store != nil
}

func (m Whitelist) HasAddress(address string) bool {
	if m.store == nil {
		return false
	}

	return m.store.Has(address)
}
