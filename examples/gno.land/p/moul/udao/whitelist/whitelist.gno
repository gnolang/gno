// Package whitelist is a simple udao implementation that allow any whitelisted address to manage proposals.
package whitelist

import (
	"std"

	"gno.land/p/moul/udao"
)

// DAO is a whitelist system that implements the udao.DAO interface.
// It can allow a trusted list of addresses to share the administration of realms that are designed to be managed by "DAOs".
// It's not a DAO per say. But makes sense to be used to bootstrap a realm until the DAO can replace itself with a really decnetralized autonomous organization. Asan inception.
// It's also useful for people who manages multiple keys, like on different computers so they can simulte being a single entity while managing several addresses.
type DAO struct {
	admins []std.Address // for the sake of simplicity and because this simple DAO implementation isn't designed to scale, we just keep a simple slice
}

// check that DAO implements the udao.DAO interface.
var _ udao.DAO = (*DAO)(nil)

func New() *DAO {
	caller := std.PrevRealm().Addr()
	return &DAO{
		admins: []std.Address{caller},
	}
}

// udao.DAO methods
//

func (d *DAO) Propose(prop udao.Proposal) (uint64, error)              { panic("not implemented") }
func (d DAO) GetProposalStatus(id uint64) (udao.ProposalStatus, error) { panic("not implemented") }
func (d *DAO) Execute(id uint64) error                                 { panic("not implemented") }
func (d DAO) GetProposal(id uint64) (udao.Proposal, error)             { panic("not implemented") }

// other methods
//

func (d *DAO) AddAdmin(addr std.Address) { panic("not implemented") }
func (d *DAO) DelAdmin(addr std.Address) { panic("not implemented") }
func (d DAO) ListAdmins() []std.Address  { return d.admins[:] }
func (d DAO) Render(path string) string  { return "TODO" }
