package todolist

import (
	"std"
	"testing"
)

func TestNewTodoList(t *testing.T) {
	title := "My Todo List"
	todoList := NewTodoList(title)

	if todoList.GetTodolistTitle() != title {
		t.Errorf("Expected title %q, got %q", title, todoList.GetTodolistTitle())
	}

	if len(todoList.GetTasks()) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(todoList.GetTasks()))
	}

	if todoList.GetTodolistOwner() != std.GetOrigCaller() {
		t.Errorf("Expected owner %v, got %v", std.GetOrigCaller(), todoList.GetTodolistOwner())
	}
}

func TestNewTask(t *testing.T) {
	title := "My Task"
	task := NewTask(title)

	if task.Title != title {
		t.Errorf("Expected title %q, got %q", title, task.Title)
	}

	if task.Done {
		t.Errorf("Expected task to be not done, but it is done")
	}
}

func TestAddTask(t *testing.T) {
	todoList := NewTodoList("My Todo List")
	task := NewTask("My Task")

	todoList.AddTask(1, task)

	tasks := todoList.GetTasks()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0] != task {
		t.Errorf("Expected task %v, got %v", task, tasks[0])
	}
}

func TestToggleTaskStatus(t *testing.T) {
	task := NewTask("My Task")

	ToggleTaskStatus(task)

	if !task.Done {
		t.Errorf("Expected task to be done, but it is not done")
	}

	ToggleTaskStatus(task)

	if task.Done {
		t.Errorf("Expected task to be not done, but it is done")
	}
}

func TestRemoveTask(t *testing.T) {
	todoList := NewTodoList("My Todo List")
	task := NewTask("My Task")
	todoList.AddTask(1, task)

	todoList.RemoveTask("1")

	tasks := todoList.GetTasks()
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}
