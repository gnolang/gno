package commondao

import (
	"std"
	"strconv"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/jeronimoalbi/pager"
	"gno.land/p/nt/commondao"
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

	var (
		dao  *commondao.CommonDAO
		path = req.Query.Get("path")
	)

	if path != "" {
		dao = getSubDAO(daoID, path)
	} else {
		dao = getDAO(daoID)
	}

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

	if parentDAO := dao.Parent(); parentDAO != nil {
		res.Write(ufmt.Sprintf("[Go to Parent DAO](%s)\n\n", daoURL(parentDAO)))
	}

	// TODO: Add link to DAO's proposals list view

	members := dao.Members()
	if members.Size() == 0 {
		res.Write("**⚠ The DAO has no members**\n\n")
	} else {
		renderMembers(res, req.RawPath, members)
	}

	if dao.Children().Len() > 0 {
		res.Write("## Tree\n")
		renderTree(res, dao, "")
	}

	// TODO: List latest 10 proposals
}

func renderMembers(res *mux.ResponseWriter, path string, members commondao.MemberStorage) {
	p, err := pager.New(path, members.Size(), pager.WithPageQueryParam("members"), pager.WithPageSize(10))
	if err != nil {
		res.Write(err.Error())
		return
	}

	res.Write("| Members |\n| --- |\n")
	members.IterateByOffset(p.Offset(), p.PageSize(), func(addr std.Address) bool {
		res.Write("| ")
		res.Write(addr.String())
		res.Write(" |\n")
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
	res.Write(ufmt.Sprintf("%s- [%s](%s)\n", indent, dao.Name(), daoURL(dao)))

	indent += "  "
	dao.Children().ForEach(func(_ int, v any) bool {
		if subDAO, ok := v.(*commondao.CommonDAO); ok {
			renderTree(res, subDAO, indent)
		}
		return false
	})
}
