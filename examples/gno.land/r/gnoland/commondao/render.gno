package commondao

import (
	"std"
	"strconv"

	"gno.land/p/demo/mux"
	"gno.land/p/gnoland/commondao"
)

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("{daoID}", renderDAO)
	return router.Render(path)
}

func renderHome(res *mux.ResponseWriter, _ *mux.Request) {
	res.Write("# Common DAO\n")

	// TODO: Render a header text for users
	// TODO: Render generic links to vote or execute proposals
}

func renderDAO(res *mux.ResponseWriter, req *mux.Request) {
	rawID := req.GetVar("daoID")
	daoID, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		res.Write("Value is an invalid DAO ID")
		return
	}

	dao := getDAO(daoID)
	if dao == nil {
		res.Write("DAO not found")
		return
	}

	res.Write("# ")
	res.Write(dao.Name())
	res.Write("\n")

	if s := dao.Description(); s != "" {
		res.Write(s)
		res.Write("\n\n")
	}

	// TODO: Add link to DAO's proposals view

	res.Write("## Members\n")

	members := dao.Members()
	if members.Size() == 0 {
		res.Write("The DAO has no members\n\n")
	} else {
		// TODO: Allow members pagination
		members.IterateByOffset(0, members.Size(), func(addr std.Address) bool {
			res.Write("- ")
			res.Write(addr.String())
			res.Write("\n")
			return false
		})
	}

	if dao.Children().Len() > 0 {
		res.Write("## Tree\n")
		renderTree(res, dao, "")
	}
}

func renderTree(res *mux.ResponseWriter, dao *commondao.CommonDAO, indent string) {
	// TODO: Use bold and no DAO link for the current DAO (requires req.Query support)
	// TODO: Render a link to the DAO page
	res.Write(indent)
	res.Write("- ")
	res.Write(dao.Name())
	res.Write("\n")

	indent += "  "
	dao.Children().ForEach(func(_ int, v any) bool {
		if subDAO, ok := v.(*commondao.CommonDAO); ok {
			renderTree(res, subDAO, indent)
		}
		return false
	})
}
