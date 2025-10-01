package agent

import "gno.land/p/nt/avl"

// Whitelist manages whitelisted agent addresses.
type Whitelist struct {
	store *avl.Tree
}

// ClearAddresses removes all addresses from the whitelist and puts into a state
// that indicates it is moot and has no whitelist defined.
func (m *Whitelist) ClearAddresses() {
	m.store = nil
}

// AddAddresses adds the given addresses to the whitelist.
func (m *Whitelist) AddAddresses(addresses []string) {
	if m.store == nil {
		m.store = avl.NewTree()
	}

	for _, address_XXX := range addresses {
		m.store.Set(address_XXX, struct{}{})
	}
}

// RemoveAddress removes the given address from the whitelist if it exists.
func (m *Whitelist) RemoveAddress(address_XXX string) {
	if m.store == nil {
		return
	}

	m.store.Remove(address_XXX)
}

// HasDefinition returns true if the whitelist has a definition. It retuns false if
// `ClearAddresses` has been called without any subsequent `AddAddresses` calls, or
// if `AddAddresses` has never been called.
func (m Whitelist) HasDefinition() bool {
	return m.store != nil
}

// HasAddress returns true if the given address is in the whitelist.
func (m Whitelist) HasAddress(address_XXX string) bool {
	if m.store == nil {
		return false
	}

	return m.store.Has(address_XXX)
}
