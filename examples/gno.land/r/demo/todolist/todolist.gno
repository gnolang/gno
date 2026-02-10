package todolistrealm

import (
	"bytes"
	"strconv"

	"gno.land/p/demo/todolist"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
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

func NewTodoList(_ realm, title string) (int, string) {
	// Create new Todolist
	tl := todolist.NewTodoList(title)
	// Update AVL tree with new state
	tlid.Next()
	todolistTree.Set(strconv.Itoa(int(tlid)), tl)
	return int(tlid), "created successfully"
}

func AddTask(_ realm, todolistID int, title string) string {
	// Get Todolist from AVL tree
	tl, ok := todolistTree.Get(strconv.Itoa(todolistID))
	if !ok {
		panic("Todolist not found")
	}

	// get the number of tasks in the todolist
	id := tl.(*todolist.TodoList).Tasks.Size()

	// create the task
	task := todolist.NewTask(title)

	// Cast raw data from tree into Todolist struct
	tl.(*todolist.TodoList).AddTask(id, task)

	return "task added successfully"
}

func ToggleTaskStatus(_ realm, todolistID int, taskID int) string {
	// Get Todolist from AVL tree
	tl, ok := todolistTree.Get(strconv.Itoa(todolistID))
	if !ok {
		panic("Todolist not found")
	}

	// Get the task from the todolist
	task, found := tl.(*todolist.TodoList).Tasks.Get(strconv.Itoa(taskID))
	if !found {
		panic("Task not found")
	}

	// Change the status of the task
	todolist.ToggleTaskStatus(task.(*todolist.Task))

	return "task status changed successfully"
}

func RemoveTask(_ realm, todolistID int, taskID int) string {
	// Get Todolist from AVL tree
	tl, ok := todolistTree.Get(strconv.Itoa(todolistID))
	if !ok {
		panic("Todolist not found")
	}

	// Get the task from the todolist
	_, ok = tl.(*todolist.TodoList).Tasks.Get(strconv.Itoa(taskID))
	if !ok {
		panic("Task not found")
	}

	// Change the status of the task
	tl.(*todolist.TodoList).RemoveTask(strconv.Itoa(taskID))

	return "task status changed successfully"
}

func RemoveTodoList(_ realm, todolistID int) string {
	// Get Todolist from AVL tree
	_, ok := todolistTree.Get(strconv.Itoa(todolistID))
	if !ok {
		panic("Todolist not found")
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
	todolistTree.Iterate("", "", func(key string, value any) bool {
		// cast raw data from tree into Todolist struct
		tl := value.(*todolist.TodoList)

		// Add Todolist name
		b.WriteString(
			ufmt.Sprintf(
				"## Todolist #%s: %s\n",
				key, // Todolist ID
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
				// displays a checked box if task is marked as done, an empty box if not
				if todo.Done {
					b.WriteString(
						"☑\n\n",
					)
					continue
				}

				b.WriteString(
					"☐\n\n",
				)
			}
		} else {
			b.WriteString("No tasks in this list currently\n")
		}

		b.WriteString("\n")
		return false
	})

	return b.String()
}
