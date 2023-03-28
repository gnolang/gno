package todo

import (
	"bytes"
	"strings"
	"std"

	"gno.land/p/demo/todo"
	"gno.land/p/demo/avl"
)

const (
	successMessage  = "New task is added successfully!"
	markMessage 		= "Good job!"
	failMessage 	  = "Unable to add the new task!"
	noTaskMessage   = "There is no task listed!"
	notFoundMessage = "Not found!"
)

var (
	admin 		std.Address = "g1m2znmyqgd32r9zcts6lp3yapmj4h832fn73n2n"
	tasks 		avl.Tree
	idCounter uint
)

func isAdmin(address std.Address) bool {
	return address == admin
}

func Render(path string) string {
	parts := strings.Split(path, "/")

	switch {
	case path == "":
		return renderTasks()
	
	case len(parts) == 2 && parts[0] == "task":
		task := getTaskById(parts[1])
		if task == nil {
			return notFoundMessage
		}

		return task.GetInfo()
	
	default:
		return notFoundMessage
	}
}

func renderTasks() string {
	if tasks.Size() < 1 {
		return noTaskMessage
	}

	var buffer bytes.Buffer

	tasks.Iterate("", "", func(t *avl.Node) bool {
		task, _ := t.Value().(*todo.Task)

		buffer.WriteString(task.GetInfo())
		buffer.WriteString("\n\n")

		return false
	})

	return buffer.String()
}

func getTaskById(id string) *todo.Task {
	task, found := task.Get(id)
	if !found {
		return nil
	}

	return task.(*todo.Task)
}

func AddTask(
	task string,
	isDone bool,
) string {
	if !isAdmin(std.GetOrigCaller()) {
		return failMessage
	}

	newTask := todo.NewTask(
		idCounter,
		task,
		isDone,
	)

	tasks.Set(
		newTask.GetIdString(),
		newTask,
	)

	idCounter++

	return successMessage
}

func MarkDone(id string) string {
	task := getTaskById(id)
	if task == nil {
		return notFoundMessage
	}

	task.isDone = true

	return markMessage
}

