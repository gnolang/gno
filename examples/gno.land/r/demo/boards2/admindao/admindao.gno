package admindao

import (
	"path"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/commondao"
)

var (
	daos     avl.Tree // string(name) -> *CommonDAO
	boardsNS string
)

func init() {
	boardsNS = path.Dir(std.CurrentRealm().PkgPath()) + "/"
}

func New(name string, options ...commondao.Option) *commondao.CommonDAO {
	assertCallerIsBoardsRealm()

	dao := commondao.New(options...)
	daos.Set(name, dao)
	return dao
}

func Get(name string) (_ *commondao.CommonDAO, found bool) {
	assertCallerIsBoardsRealm()

	v, found := daos.Get(name)
	if !found {
		return nil, false
	}
	return v.(*commondao.CommonDAO), true
}

func Exists(name string) bool {
	return daos.Has(name)
}

func assertCallerIsBoardsRealm() {
	if !strings.HasPrefix(std.PrevRealm().PkgPath(), boardsNS) {
		panic("unauthorized")
	}
}
