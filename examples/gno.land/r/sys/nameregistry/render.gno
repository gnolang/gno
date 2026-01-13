package nameregistry

import (
	"chain/runtime"
	"strconv"
	"strings"

	"gno.land/p/jeronimoalbi/mdform"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/mdalert"
	"gno.land/p/nt/mux"
)

const pathVerify = "verify"

var realmPath = strings.TrimPrefix(runtime.CurrentRealm().PkgPath(), "gno.land")

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc(pathVerify, renderVerify)

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
		var b strings.Builder
		for _, c := range constraints {
			if err := c.Check(name); err != nil {
				b.WriteString(md.Paragraph(err.Error()))
			}
		}

		messages := b.String()
		if messages != "" {
			res.Write(mdalert.Caution("Name is not valid", messages))
		} else {
			res.Write(mdalert.Info("Info", "The name is a valid"))
		}
	}

	res.Write(form.String())
}

func makeURI(path string) string {
	if path == "" {
		return realmPath
	}
	return realmPath + ":" + path
}
