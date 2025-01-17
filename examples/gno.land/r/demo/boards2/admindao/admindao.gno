package admindao

import (
	"path"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/boards2/admindao"
)

var (
	daos     avl.Tree // string(name) -> *AdminDAO
	boardsNS string
)

func init() {
	boardsNS = path.Dir(std.CurrentRealm().PkgPath()) + "/"
}

func New(name string, options ...admindao.Option) *admindao.AdminDAO {
	assertCallerIsBoardsRealm()

	dao := admindao.New(options...)
	daos.Set(name, dao)
	return dao
}

func Get(name string) (_ *admindao.AdminDAO, found bool) {
	assertCallerIsBoardsRealm()

	v, found := daos.Get(name)
	if !found {
		return nil, false
	}
	return v.(*admindao.AdminDAO), true
}

func Exists(name string) bool {
	return daos.Has(name)
}

func assertCallerIsBoardsRealm() {
	if !strings.HasPrefix(std.PrevRealm().PkgPath(), boardsNS) {
		panic("unauthorized")
	}
}
