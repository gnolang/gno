package todolist

import (
	"testing"
)

func TestNewTodoList(t *testing.T) {
	title := "My Todo List"
	todoList := NewTodoList(title)

	if todoList.GetTodolistTitle() != title {
		t.Errorf("Expected title to be %s, but got %s", title, todoList.GetTodolistTitle())
	}

	if todoList.GetTodolistOwner() != std.GetOrigCaller() {
		t.Errorf("Expected owner to be %s, but got %s", std.GetOrigCaller(), todoList.GetTodolistOwner())
	}

	if len(todoList.GetTasks()) != 0 {
		t.Errorf("Expected no tasks in the todo list, but got %d tasks", len(todoList.GetTasks()))
	}
}

func TestAddTask(t *testing.T) {
	todoList := NewTodoList("My Todo List")
	task := NewTask("Task 1")

	todoList.AddTask(1, task)

	tasks := todoList.GetTasks()
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
	task := NewTask("Task 1")

	ToggleTaskStatus(task)

	if !task.Done {
		t.Errorf("Expected task to be done, but it is not marked as done")
	}

	ToggleTaskStatus(task)

	if task.Done {
		t.Errorf("Expected task to be not done, but it is marked as done")
	}
}

func TestRemoveTask(t *testing.T) {
	todoList := NewTodoList("My Todo List")
	task := NewTask("Task 1")

	todoList.AddTask(1, task)

	todoList.RemoveTask("1")

	tasks := todoList.GetTasks()
	if len(tasks) != 0 {
		t.Errorf("Expected no tasks in the todo list, but got %d tasks", len(tasks))
	}
}
