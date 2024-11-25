package dao

import "gno.land/p/moul/xdao"

var DAO = xdao.NewWhitelistDAO()

func Render(path string) string {
	return DAO.Render(path)
}
