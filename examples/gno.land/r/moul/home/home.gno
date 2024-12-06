package home

import (
	"strconv"

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
	content += umd.Paragraph(
		umd.Image("meme", memeImgURL),
	)

	content += umd.H2("Status")
	content += umd.Paragraph(status)
	content += umd.Link("update", txlink.Call("UpdateStatus"))

	content += umd.H2("Personal ToDo List")
	content += umd.BulletItem("test")
	for i, todo := range todos {
		deleteLink := umd.Link("del", txlink.Call("DeleteTodo", "idx", strconv.Itoa(i)))
		content += umd.BulletItem(todo + " " + deleteLink)
	}
	content += umd.Paragraph(
		umd.Link("new todo", txlink.Call("AddNewTodo")),
	)

	// TODO: my r/boards posts
	// TODO: my r/events events
	return content
}

func AddNewTodo(todo string) {
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
