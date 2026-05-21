package todolistrealm

import (
	"chain/runtime"
	"strconv"
	"testing"

	"gno.land/p/demo/todolist"
	"gno.land/p/nt/uassert/v0"
)

func setupTodoList(t *testing.T, title string) (int, *todolist.TodoList) {
	t.Helper()
	tlid, _ := NewTodoList(cross, title)
	node, _ := todolistTree.Get(strconv.Itoa(tlid))
	return tlid, node.(*todolist.TodoList)
}

func TestNewTodoList(t *testing.T) {
	title := "My Todo List"
	tlid, tdl := setupTodoList(t, title)

	uassert.Equal(t, 1, tlid, "tlid does not match")
	uassert.Equal(t, title, tdl.Title, "title does not match")
	uassert.Equal(t, tdl.Owner.String(), runtime.OriginCaller().String(), "owner does not match")
	uassert.Equal(t, 0, len(tdl.GetTasks()), "Expected no tasks in the todo list")
}

func TestAddTask(t *testing.T) {
	tlid, tdl := setupTodoList(t, "My Todo List")
	AddTask(cross, tlid, "Task 1")

	tasks := tdl.GetTasks()
	uassert.Equal(t, 1, len(tasks), "total task does not match")
	uassert.Equal(t, "Task 1", tasks[0].Title, "task title does not match")
	uassert.False(t, tasks[0].Done, "Expected task to be not done")
}

func TestToggleTaskStatus(t *testing.T) {
	tlid, tdl := setupTodoList(t, "My Todo List")
	AddTask(cross, tlid, "Task 1")

	ToggleTaskStatus(cross, tlid, 0)
	task := tdl.GetTasks()[0]
	uassert.True(t, task.Done, "Expected task to be done, but it is not marked as done")

	ToggleTaskStatus(cross, tlid, 0)
	uassert.False(t, task.Done, "Expected task to be not done, but it is marked as done")
}

func TestRemoveTask(t *testing.T) {
	tlid, tdl := setupTodoList(t, "My Todo List")
	AddTask(cross, tlid, "Task 1")

	RemoveTask(cross, tlid, 0)
	uassert.Equal(t, 0, len(tdl.GetTasks()), "Expected no tasks in the todo list")
}

func TestRemoveTodoList(t *testing.T) {
	tlid, _ := setupTodoList(t, "My Todo List")

	RemoveTodoList(cross, tlid)
	uassert.Equal(t, 0, todolistTree.Size(), "Expected no todolists after removal")
}
