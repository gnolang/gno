package rcdao

import (
	"gno.land/p/nt/commondao"
)

var dao *commondao.CommonDAO

func init() {
	dao = commondao.New(commondao.WithName("Regional Community DAO"))
}

func Render(_ string) string {

}
