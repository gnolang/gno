package todolist

import (
	"bytes"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/todolist"
	"gno.land/p/demo/ufmt"
)

// State variables
var (
	todolistTree *avl.Tree
	tlid         seqid.ID
)

// Constructor
func init() {
	todolistTree = avl.NewTree()
}

func NewTodolist(title string) (int, string) {
	// Create new Todolist
	tl := todolist.NewTodoList(title)
	// Update AVL tree with new state
	todolistTree.Set(tlid.String(), tl)
	tlid.Next()
	return tlid - 1, "created successfully"
}

func AddTask(todolistID int, title string) string {
	// Get Todolist from AVL tree
	tl, ok := todolistTree.Get(strconv.Itoa(todolistID))
	if !ok {
		return "Todolist not found"
	}

	// get the number of tasks in the todolist
	id := tl.(*todolist.TodoList).Tasks.Size()

	// create the task
	task := todolist.NewTask(title)

	// Cast raw data from tree into Todolist struct
	tl.(*todolist.TodoList).AddTask(id, task)

	return "task added successfully"
}

func ToggleTaskStatus(todolistID int, taskID int) string {
	// Get Todolist from AVL tree
	tl, ok := todolistTree.Get(strconv.Itoa(todolistID))
	if !ok {
		return "Todolist not found"
	}

	// Get the task from the todolist
	task, found := tl.(*todolist.TodoList).Tasks.Get(strconv.Itoa(taskID))
	if !found {
		return "Task not found"
	}

	// Change the status of the task
	todolist.ToggleTaskStatus(task.(*todolist.Task))

	return "task status changed successfully"
}

func RemoveTask(todolistID int, taskID int) string {
	// Get user who sent the transaction
	// txSender := std.GetOrigCaller()

	// Get Todolist from AVL tree
	tl, ok := todolistTree.Get(strconv.Itoa(todolistID))
	if !ok {
		return "Todolist not found"
	}

	// Get the task from the todolist
	_, ok = tl.(*todolist.TodoList).Tasks.Get(strconv.Itoa(taskID))
	if !ok {
		return "Task not found"
	}

	// Change the status of the task
	tl.(*todolist.TodoList).RemoveTask(strconv.Itoa(taskID))

	return "task status changed successfully"
}

func RemoveTodoList(todolistID int) string {
	// Get user who sent the transaction
	// txSender := std.GetOrigCaller()

	// Get Todolist from AVL tree
	_, ok := todolistTree.Get(strconv.Itoa(todolistID))
	if !ok {
		return "Todolist not found"
	}

	// Remove the todolist
	todolistTree.Remove(strconv.Itoa(todolistID))

	return "Todolist removed successfully"
}

func Render(path string) string {
	if path == "" {
		return renderHomepage()
	}

	return "unknown page"
}

func renderHomepage() string {
	// Define empty buffer
	var b bytes.Buffer

	b.WriteString("# Welcome to ToDolist\n\n")

	// If no todolists have been created
	if todolistTree.Size() == 0 {
		b.WriteString("### No todolists available currently!")
		return b.String()
	}

	// Iterate through AVL tree
	todolistTree.Iterate("", "", func(key string, value interface{}) bool {
		// cast raw data from tree into Todolist struct
		tl := value.(*todolist.TodoList)

		// Add Todolist name
		b.WriteString(
			ufmt.Sprintf(
				"## Todolist #%s: %s\n",
				key+1, // Todolist ID
				tl.GetTodolistTitle(),
			),
		)

		// Add Todolist owner
		b.WriteString(
			ufmt.Sprintf(
				"#### Todolist owner : %s\n",
				tl.GetTodolistOwner(),
			),
		)

		// List all todos that are currently Todolisted
		if todos := tl.GetTasks(); len(todos) > 0 {
			b.WriteString(
				ufmt.Sprintf("Currently Todo tasks: %d\n\n", len(todos)),
			)

			for index, todo := range todos {
				b.WriteString(
					ufmt.Sprintf("#%d - %s  ", index, todo.Title),
				)
				if todo.Done {
					b.WriteString(
						"☑\n\n",
					)
				} else {
					b.WriteString(
						"☐\n\n",
					)
				}
			}
		} else {
			b.WriteString("No tasks in this list currently\n")
		}

		b.WriteString("\n")
		return false
	})

	return b.String()
}
