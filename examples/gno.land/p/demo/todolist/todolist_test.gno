package todolist

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestNewTodoList(t *testing.T) {
	title := "My Todo List"
	todoList := NewTodoList(title)

	uassert.Equal(t, title, todoList.GetTodolistTitle())
	uassert.Equal(t, 0, len(todoList.GetTasks()))
	uassert.Equal(t, runtime.OriginCaller().String(), todoList.GetTodolistOwner().String())
}

func TestNewTask(t *testing.T) {
	title := "My Task"
	task := NewTask(title)

	uassert.Equal(t, title, task.Title)
	uassert.False(t, task.Done, "Expected task to be not done, but it is done")
}

func TestAddTask(t *testing.T) {
	todoList := NewTodoList("My Todo List")
	task := NewTask("My Task")

	todoList.AddTask(1, task)

	tasks := todoList.GetTasks()

	uassert.Equal(t, 1, len(tasks))
	uassert.True(t, tasks[0] == task, "Task does not match")
}

func TestToggleTaskStatus(t *testing.T) {
	task := NewTask("My Task")

	ToggleTaskStatus(task)
	uassert.True(t, task.Done, "Expected task to be done, but it is not done")

	ToggleTaskStatus(task)
	uassert.False(t, task.Done, "Expected task to be done, but it is not done")
}

func TestRemoveTask(t *testing.T) {
	todoList := NewTodoList("My Todo List")
	task := NewTask("My Task")
	todoList.AddTask(1, task)

	todoList.RemoveTask("1")

	tasks := todoList.GetTasks()
	uassert.Equal(t, 0, len(tasks))
}
