package nameregistry

import (
	"chain/runtime"
	"strconv"
	"strings"

	"gno.land/p/jefft0/avlhelpers"
	"gno.land/p/jeronimoalbi/mdform"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/mdalert"
	"gno.land/p/nt/mux"
)

const (
	pathVerify = "verify"
	pathSearch = "search"
)

const maxSearchItems = 30

var realmPath = strings.TrimPrefix(runtime.CurrentRealm().PkgPath(), "gno.land")

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc(pathVerify, renderVerify)
	router.HandleFunc(pathSearch, renderSearch)

	router.NotFoundHandler = func(res *mux.ResponseWriter, _ *mux.Request) {
		res.Write(md.Blockquote("Path not found"))
	}

	var b strings.Builder
	b.WriteString(md.H1("Name Registry"))
	b.WriteString(router.Render(path))
	return b.String()
}

func renderHome(res *mux.ResponseWriter, _ *mux.Request) {
	items := make([]string, len(constraints))
	for i, c := range constraints {
		items[i] = c.Description()
	}

	table := mdtable.Table{
		Headers: []string{"Names", "Total"},
		Rows: [][]string{
			{"Reserved", strconv.Itoa(reserved.Size())},
			{"Blacklisted", strconv.Itoa(blacklisted.Size())},
			{"Custom", strconv.Itoa(custom.Size())},
		},
	}

	res.Write(" ↳ ")
	res.Write(md.Link("Verify Name", makeURI(pathVerify)))
	res.Write(" • ")
	res.Write(md.Link("Search", makeURI(pathSearch)))
	res.Write("\n")

	res.Write(md.H2("Naming Constraints"))
	res.Write(md.Paragraph("Names must match the following constraints:"))
	res.Write(md.OrderedList(items))

	res.Write(md.H2("Registered Names"))
	res.Write(table.String())
}

func renderVerify(res *mux.ResponseWriter, req *mux.Request) {
	name := strings.TrimSpace(req.Query.Get("name"))
	form := mdform.New("path", pathVerify)
	form.Input(
		"name",
		"placeholder", "Write a name to verify",
		"value", name,
	)

	res.Write(md.Link("← Back Home", makeURI("")) + "\n\n")
	res.Write(md.H2("Verify Name"))
	res.Write(
		md.Paragraph(
			"Verification checks if a name is valid for current registry contraints",
		),
	)

	if name != "" {
		var errors strings.Builder
		for _, c := range constraints {
			if err := c.Check(name); err != nil {
				errors.WriteString(md.Paragraph(err.Error()))
			}
		}

		if errors.Len() == 0 {
			res.Write(mdalert.Info("Info", "The name is a valid"))
		} else {
			res.Write(mdalert.Caution("Name is not valid", errors.String()))
		}
	}

	res.Write(form.String())
}

func renderSearch(res *mux.ResponseWriter, req *mux.Request) {
	nameType := req.Query.Get("type")
	prefix := strings.TrimSpace(req.Query.Get("prefix"))

	form := mdform.New("path", pathSearch)
	form.Input(
		"prefix",
		"placeholder", "Name Prefix",
		"value", prefix,
	)
	form.Select(
		"type",
		"Reserved",
		"selected", strconv.FormatBool(nameType == "" || nameType == "Reserved"),
	)
	form.Select(
		"type",
		"Custom",
		"selected", strconv.FormatBool(nameType == "Custom"),
	)
	form.Select(
		"type",
		"Blacklisted",
		"selected", strconv.FormatBool(nameType == "Blacklisted"),
	)

	var names *avl.Tree
	switch nameType {
	case "Reserved":
		names = reserved
	case "Custom":
		names = custom
	case "Blacklisted":
		names = blacklisted
	}

	res.Write(md.Link("← Back Home", makeURI("")) + "\n\n")
	res.Write(md.H2("Search Names"))

	if names == nil || prefix == "" {
		res.Write(form.String())
	} else {
		var (
			results strings.Builder
			current int
		)

		avlhelpers.IterateByteStringKeysByPrefix(names, prefix, func(name string, _ any) bool {
			results.WriteString(md.BulletItem(name))
			current++
			return current == maxSearchItems
		})

		if results.Len() == 0 {
			alert := mdalert.New(
				mdalert.TypeWarning,
				"Not Found",
				"Search didn't find any matching names",
				false,
			)
			res.Write(alert.String())
			res.Write(form.String())
		} else {
			res.Write(form.String())
			res.Write(results.String())
		}
	}
}

func makeURI(path string) string {
	if path == "" {
		return realmPath
	}
	return realmPath + ":" + path
}
