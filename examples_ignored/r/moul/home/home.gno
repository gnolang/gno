package home

import (
	"strconv"

	"gno.land/p/demo/svg"
	"gno.land/p/moul/debug"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/web25"
	"gno.land/r/leon/hor"
	"gno.land/r/moul/config"
)

var (
	todos       []string
	status      string
	memeImgURL  string
	web25config = web25.Config{URL: "https://moul.github.io/gno-moul-home-web25/"}
)

func init() {
	todos = append(todos, "fill this todo list...")
	status = "Online" // Initial status set to "Online"
	memeImgURL = "https://i.imgflip.com/7ze8dc.jpg"
	hor.Register("Moul's Home Realm!", "")
}

func Render(path string) string {
	content := web25config.Render(path)
	var d debug.Debug

	content += md.H1("Manfred's (gn)home Dashboard")

	content += md.H2("Meme")
	content += md.Paragraph(
		md.Image("meme", memeImgURL),
	)

	content += md.H2("Status")
	content += md.Paragraph(status)
	content += md.Paragraph(md.Link("update", txlink.Call("UpdateStatus")))

	d.Log("hello world!")

	content += md.H2("Personal TODO List (bullet list)")
	for i, todo := range todos {
		idstr := strconv.Itoa(i)
		deleteLink := md.Link("x", txlink.Call("DeleteTodo", "idx", idstr))
		content += md.BulletItem(todo + " " + deleteLink)
	}
	content += md.BulletItem(md.Link("[new]", txlink.Call("AddTodo")))

	content += md.H2("Personal TODO List (table)")
	table := mdtable.Table{
		Headers: []string{"ID", "Item", "Links"},
	}
	for i, todo := range todos {
		idstr := strconv.Itoa(i)
		deleteLink := md.Link("[del]", txlink.Call("DeleteTodo", "idx", idstr))
		table.Append([]string{"#" + idstr, todo, deleteLink})
	}
	content += table.String()

	content += md.H2("SVG Example")
	content += md.Paragraph("this feature may not work with the current gnoweb version and/or configuration.")
	content += md.Paragraph(svg.Canvas{
		Width: 500, Height: 500,
		Elems: []svg.Elem{
			svg.Rectangle{50, 50, 100, 100, "red"},
			svg.Circle{50, 50, 100, "red"},
			svg.Text{100, 100, "hello world!", "magenta"},
		},
	}.String())

	content += md.H2("Debug")
	content += md.Paragraph("this feature may not work with the current gnoweb version and/or configuration.")
	content += md.Paragraph(
		md.Link("toggle debug", debug.ToggleURL(path)),
	)

	// TODO: my r/boards posts
	// TODO: my r/events events
	content += d.Render(path)
	return content
}

func AddTodo(todo string) {
	config.AssertIsAdmin()
	todos = append(todos, todo)
}

func DeleteTodo(idx int) {
	config.AssertIsAdmin()
	if idx >= 0 && idx < len(todos) {
		// Remove the todo from the list by merging slices from before and after the todo
		todos = append(todos[:idx], todos[idx+1:]...)
	} else {
		panic("Invalid todo index")
	}
}

func UpdateStatus(newStatus string) {
	config.AssertIsAdmin()
	status = newStatus
}
