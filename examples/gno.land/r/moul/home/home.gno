package home

import (
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/umd"
	"gno.land/p/moul/web25"
	"gno.land/r/leon/hof"
	"gno.land/r/moul/config"
)

var (
	todos       []string
	status      string
	memeImgURL  string
	web25config = web25.Config{CID: "12345678"}
)

func init() {
	todos = append(todos, "fill this todo list...")
	status = "Online" // Initial status set to "Online"
	memeImgURL = "https://i.imgflip.com/7ze8dc.jpg"
	hof.Register()
}

func Render(path string) string {
	content := web25config.Render(path)

	content += umd.H1("Manfred's (gn)home Dashboard")

	content += umd.H2("Meme")
	content += umd.Paragraph(umd.Image("meme", memeImgURL))

	content += umd.H2("Status")
	content += umd.Paragraph(status)

	content += umd.H2("Personal ToDo List")
	for _, todo := range todos {
		content += "- [ ] " + todo + "\n"
	}
	content += txlink.URL("foo")

	// TODO: Implement a feature to list replies on r/boards on my posts
	// TODO: Maybe integrate a calendar feature for upcoming events?

	return content
}

func AddNewTodo(todo string) {
	config.AssertIsAdmin()
	todos = append(todos, todo)
}

func DeleteTodo(todoIndex int) {
	config.AssertIsAdmin()
	if todoIndex >= 0 && todoIndex < len(todos) {
		// Remove the todo from the list by merging slices from before and after the todo
		todos = append(todos[:todoIndex], todos[todoIndex+1:]...)
	} else {
		panic("Invalid todo index")
	}
}

func UpdateStatus(newStatus string) {
	config.AssertIsAdmin()
	status = newStatus
}
