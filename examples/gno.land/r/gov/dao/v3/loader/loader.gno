package loader

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
)

// this is only executed when loaded into genesis
func init() {
	cross(dao.UpdateImpl)(dao.UpdateRequest{
		DAO:         impl.GetInstance(),
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
}
