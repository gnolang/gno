package rcdao

import (
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"gno.land/p/nt/commondao"
)

var dao *commondao.CommonDAO

func init() {
	dao = commondao.New(
		commondao.WithName("Regional Community DAO"),
		commondao.WithMember("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"))
}

// I want to have a quick view of members
// I want to have an easy render, name it DefaultRender
// Exposing RO trees for members & props would make it easy to integrate with existing p/

func Render(_ string) string {

	out := "# Regional Community DAO\n"

	dao.ActiveProposals().Iterate(0, 0, false, func(prop *commondao.Proposal) bool {
		out += prop.CreatedAt()
		return false
	})
}
