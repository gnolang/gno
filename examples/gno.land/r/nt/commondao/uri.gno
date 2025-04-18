package commondao

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/nt/commondao"
)

// TODO: Improve URL functions to remove redundancy

func currentRealmPath() string {
	return strings.TrimPrefix(std.CurrentRealm().PkgPath(), std.ChainDomain())
}

func daoURL(dao *commondao.CommonDAO) string {
	path := currentRealmPath()
	if dao.Parent() == nil {
		return ufmt.Sprintf("%s:%d", path, dao.ID())
	}

	rootDAO := dao.TopParent()
	return ufmt.Sprintf("%s:%d?path=%s", path, rootDAO.ID(), dao.Path())
}

func daoProposalsURL(dao *commondao.CommonDAO) string {
	path := currentRealmPath()
	if dao.Parent() == nil {
		return ufmt.Sprintf("%s:%d/proposals", path, dao.ID())
	}

	rootDAO := dao.TopParent()
	return ufmt.Sprintf("%s:%d/proposals?path=%s", path, rootDAO.ID(), dao.Path())
}

func proposalURL(dao *commondao.CommonDAO, p *commondao.Proposal) string {
	path := currentRealmPath()
	if dao.Parent() == nil {
		return ufmt.Sprintf("%s:%d/proposals/%d", path, dao.ID(), p.ID())
	}

	rootDAO := dao.TopParent()
	return ufmt.Sprintf("%s:%d/proposals/%d?path=%s", path, rootDAO.ID(), p.ID(), dao.Path())
}

func voteURL(dao *commondao.CommonDAO, p *commondao.Proposal, addr std.Address) string {
	path := currentRealmPath()
	if dao.Parent() == nil {
		return ufmt.Sprintf("%s:%d/proposals/%d/vote/%s", path, dao.ID(), p.ID(), addr)
	}

	rootDAO := dao.TopParent()
	return ufmt.Sprintf("%s:%d/proposals/%d/vote/%s?path=%s", path, rootDAO.ID(), p.ID(), dao.Path(), addr)
}
