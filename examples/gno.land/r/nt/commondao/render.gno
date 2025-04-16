package commondao

import (
	"std"
	"strconv"

	"gno.land/p/demo/mux"
	"gno.land/p/gnoland/commondao"
	"gno.land/p/jeronimoalbi/pager"
)

func Render(path string) string {
	// TODO: Add view to list proposals of a DAO
	// TODO: Add view to render a single proposal (allow/disallow must be configurable)
	// TODO: Add view to render a SubDAO (use query "?path=...")
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("{daoID}", renderDAO)
	return router.Render(path)
}

func renderHome(res *mux.ResponseWriter, _ *mux.Request) {
	// TODO: Render a header text for users
	// TODO: Render generic links to vote or execute proposals
	res.Write("# Common DAO\n")
}

func renderDAO(res *mux.ResponseWriter, req *mux.Request) {
	// TODO: Support config option to allow/disallow rendering a DAO
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

	// TODO: Add link to DAO's proposals list view

	res.Write("## Members\n")
	renderMembers(res, dao, req.RawPath)

	if dao.Children().Len() > 0 {
		res.Write("## Tree\n")
		renderTree(res, dao, "")
	}

	// TODO: List latest 10 proposals
}

func renderMembers(res *mux.ResponseWriter, dao *commondao.CommonDAO, path string) {
	members := dao.Members()
	membersCount := members.Size()
	if membersCount == 0 {
		res.Write("The DAO has no members\n\n")
		return
	}

	p, err := pager.New(path, membersCount, pager.WithPageQueryParam("members"), pager.WithPageSize(10))
	if err != nil {
		res.Write(err.Error())
		return
	}

	members.IterateByOffset(p.Offset(), p.PageSize(), func(addr std.Address) bool {
		res.Write("- ")
		res.Write(addr.String())
		res.Write("\n")
		return false
	})

	if p.HasPages() {
		res.Write("\n")
		res.Write(pager.Picker(p))
		res.Write("\n")
	}

	res.Write("\n")
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
