package dao

import "gno.land/p/moul/udao/whitelist"

var DAO = whitelist.New()

func Render(path string) string {
	return DAO.Render(path)
}
