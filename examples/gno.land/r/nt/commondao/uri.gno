package commondao

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/nt/commondao"
)

func currentRealmPath() string {
	return strings.TrimPrefix(string(realmLink), std.ChainDomain())
}

func daoURL(dao *commondao.CommonDAO) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d", path, dao.ID())
}

func settingsURL(dao *commondao.CommonDAO) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d/settings", path, dao.ID())
}

func daoProposalsURL(dao *commondao.CommonDAO) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d/proposals", path, dao.ID())
}

func proposalURL(dao *commondao.CommonDAO, p *commondao.Proposal) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d/proposals/%d", path, dao.ID(), p.ID())
}

func voteURL(dao *commondao.CommonDAO, p *commondao.Proposal, addr std.Address) string {
	path := currentRealmPath()
	return ufmt.Sprintf("%s:%d/proposals/%d/vote/%s", path, dao.ID(), p.ID(), addr)
}
