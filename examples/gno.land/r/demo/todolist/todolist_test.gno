package todolistrealm

import (
	"chain/runtime"
	"strconv"
	"testing"

	"gno.land/p/demo/todolist"
	"gno.land/p/nt/uassert/v0"
)

var (
	node any
	tdl  *todolist.TodoList
)

func TestNewTodoList(t *testing.T) {
	title := "My Todo List"
	tlid, _ := NewTodoList(cross, title)
	uassert.Equal(t, 1, tlid, "tlid does not match")

	// get the todolist node from the tree
	node, _ = todolistTree.Get(strconv.Itoa(tlid))
	// convert the node to a TodoList struct
	tdl = node.(*todolist.TodoList)

	uassert.Equal(t, title, tdl.Title, "title does not match")
	uassert.Equal(t, 1, tlid, "tlid does not match")
	uassert.Equal(t, tdl.Owner.String(), runtime.OriginCaller().String(), "owner does not match")
	uassert.Equal(t, 0, len(tdl.GetTasks()), "Expected no tasks in the todo list")
}

func TestAddTask(t *testing.T) {
	AddTask(cross, 1, "Task 1")

	tasks := tdl.GetTasks()
	uassert.Equal(t, 1, len(tasks), "total task does not match")
	uassert.Equal(t, "Task 1", tasks[0].Title, "task title does not match")
	uassert.False(t, tasks[0].Done, "Expected task to be not done")
}

func TestToggleTaskStatus(t *testing.T) {
	ToggleTaskStatus(cross, 1, 0)
	task := tdl.GetTasks()[0]
	uassert.True(t, task.Done, "Expected task to be done, but it is not marked as done")

	ToggleTaskStatus(cross, 1, 0)
	uassert.False(t, task.Done, "Expected task to be not done, but it is marked as done")
}

func TestRemoveTask(t *testing.T) {
	RemoveTask(cross, 1, 0)
	tasks := tdl.GetTasks()
	uassert.Equal(t, 0, len(tasks), "Expected no tasks in the todo list")
}

func TestRemoveTodoList(t *testing.T) {
	RemoveTodoList(cross, 1)
	uassert.Equal(t, 0, todolistTree.Size(), "Expected no tasks in the todo list")
}
