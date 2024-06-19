package home

import "gno.land/r/manfred/config"

var (
	todos      []string
	status     string
	memeImgURL string
)

func init() {
	todos = append(todos, "fill this todo list...")
	status = "Online" // Initial status set to "Online"
	memeImgURL = "https://i.imgflip.com/7ze8dc.jpg"
}

func Render(path string) string {
	content := "# Manfred's (gn)home Dashboard\n\n"

	content += "## Meme\n"
	content += "![](" + memeImgURL + ")\n\n"

	content += "## Status\n"
	content += status + "\n\n"

	content += "## Personal ToDo List\n"
	for _, todo := range todos {
		content += "- [ ] " + todo + "\n"
	}
	content += "\n"

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
