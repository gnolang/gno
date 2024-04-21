package todolistrealm

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/todolist"
)

var (
	node interface{}
	tdl  *todolist.TodoList
)

func TestNewTodoList(t *testing.T) {
	title := "My Todo List"
	tlid, _ := NewTodoList(title)
	if tlid != 1 {
		t.Errorf("Expected tlid to be 1, but got %d", tlid)
	}

	// get the todolist node from the tree
	node, _ = todolistTree.Get(strconv.Itoa(tlid))
	// convert the node to a TodoList struct
	tdl = node.(*todolist.TodoList)

	if tdl.Title != title {
		t.Errorf("Expected title to be %s, but got %s", title, tdl.Title)
	}
	if tdl.Owner != std.GetOrigCaller() {
		t.Errorf("Expected owner to be %s, but got %s", std.GetOrigCaller(), tdl.Owner)
	}
	if len(tdl.GetTasks()) != 0 {
		t.Errorf("Expected no tasks in the todo list, but got %d tasks", len(tdl.GetTasks()))
	}
}

func TestAddTask(t *testing.T) {
	AddTask(1, "Task 1")

	tasks := tdl.GetTasks()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task in the todo list, but got %d tasks", len(tasks))
	}

	if tasks[0].Title != "Task 1" {
		t.Errorf("Expected task title to be 'Task 1', but got '%s'", tasks[0].Title)
	}

	if tasks[0].Done {
		t.Errorf("Expected task to be not done, but it is marked as done")
	}
}

func TestToggleTaskStatus(t *testing.T) {
	ToggleTaskStatus(1, 0)
	task := tdl.GetTasks()[0]

	if !task.Done {
		t.Errorf("Expected task to be done, but it is not marked as done")
	}

	ToggleTaskStatus(1, 0)

	if task.Done {
		t.Errorf("Expected task to be not done, but it is marked as done")
	}
}

func TestRemoveTask(t *testing.T) {
	RemoveTask(1, 0)
	tasks := tdl.GetTasks()

	if len(tasks) != 0 {
		t.Errorf("Expected no tasks in the todo list, but got %d tasks", len(tasks))
	}
}

func TestRemoveTodoList(t *testing.T) {
	RemoveTodoList(1)

	if todolistTree.Size() != 0 {
		t.Errorf("Expected no tasks in the todo list, but got %d tasks", todolistTree.Size())
	}
}
