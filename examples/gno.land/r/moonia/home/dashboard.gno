package home

import (
	// "gno.land/p/moonia/utils"
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/md"
)

func renderDashboardDesc() string {
	out := "Welcome to the platform that lists all active DAOs on Moonia.\n\n"
	out += "🔍 You can explore each DAO’s activities and proposals freely.\n\n"
	out += "🗳️ To **participate in votes or submit proposals**, you must request to join the DAO.\n"
	out += "The admin will review and accept or decline your request.\n\n"
	return out
}

func renderDashboard() string {
	// caller := utils.GetCaller()
	out := "# Welcome to the DAO Hub\n\n"
	out += renderDashboardDesc()

	out += "## ➕ Create a new DAO:\n"
	out += md.Link("Create DAO", txlink.Call("CreateDAO", "args", "id", "args", "title", "args", "description")) + "\n\n"

	out += "## 🗂 Existing DAOs:\n"
	if len(daoMap) == 0 {
		out += "_No DAOs created yet._\n"
	} else {
		for id, dao := range daoMap {
			out += "- " + md.Link(dao.Name, "/r/moonia/home?dao="+id) + " : " + dao.Description
		}		
	}
	out += "\n## 👥 My DAOs:\n"
	// TODO: display addr DAOs
	out += "_You are not a member of any DAO._\n"
	return out
}
