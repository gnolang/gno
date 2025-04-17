package commondao

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/nt/commondao"
)

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
