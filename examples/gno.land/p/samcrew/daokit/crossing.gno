package daokit

// Enables DAOs to work across multiple realms.
// It wraps all operations that modify state (Vote, Execute, Propose).
// Read operations (Extension, Render) pass through directly.
//
// Example: A governance DAO managing both treasury and marketplace modules.
//
// Usage:
//   crossingDAO := NewCrossing(localDAO, crossFn)
//   crossingDAO.Vote(proposalID, daocond.VoteYes) // Works across realms

import "gno.land/p/samcrew/daocond"

type CrossFn func(cur realm, cb func())

// DAO wrapper that handles cross-realm communication
type crossingDAO struct {
	dao     DAO     // The underlying DAO implementation
	crossFn CrossFn // Function for cross-realm calls
}

// Creates a new DAO wrapper that enables cross-realm communication.
func NewCrossing(dao DAO, crossFn CrossFn) DAO {
	return &crossingDAO{dao: dao, crossFn: crossFn}
}

// Wraps voting operations with cross-realm functionality.
func (c *crossingDAO) Vote(proposalID uint64, vote daocond.Vote) {
	c.crossFn(cross, func() {
		c.dao.Vote(proposalID, vote)
	})
}

// Wraps proposal execution with cross-realm functionality.
func (c *crossingDAO) Execute(proposalID uint64) {
	c.crossFn(cross, func() {
		c.dao.Execute(proposalID)
	})
}

// Wraps proposal creation with cross-realm functionality and returns the proposal ID.
func (c *crossingDAO) Propose(req ProposalRequest) uint64 {
	var id uint64
	c.crossFn(cross, func() {
		id = c.dao.Propose(req)
	})
	return id
}

func (c *crossingDAO) Extension(path string) Extension {
	return c.dao.Extension(path)
}

func (c *crossingDAO) ExtensionsList() ExtensionsList {
	return c.dao.ExtensionsList()
}

func (c *crossingDAO) Render(path string) string {
	return c.Render(path)
}

var _ DAO = (*crossingDAO)(nil)
