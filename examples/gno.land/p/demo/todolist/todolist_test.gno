package todolist

import (
	"testing"

	"gno.land/p/demo/avl"
)

func TestTodolistNew(t *testing.T) {
	newtl := &TodoList{
		Title: "new todolist",
		Tasks: avl.Tree{},
	}
	tl := NewTodoList("new todolist")
	if tl.Title != newtl.Title {
		t.Fatalf("title is not good")
	}
	if tl.Tasks != newtl.Tasks {
		t.Fatalf("tasks not good")
	}

	// test new task
	newtask := &Task{
		Title: "new task",
		Done:  false,
	}
	nt := NewTask("new task")
	if nt.Title != newtask.Title {
		t.Fatalf("title is not good")
	}
	if nt.Done != newtask.Done {
		t.Fatalf("done is not good")
	}

	// test add a task
	id := tl.Tasks.Size()
	tl.AddTask(id, nt)
	if tl.Tasks.Size() != 1 {
		t.Fatalf("task not added")
	}

	// test ToggleTaskStatus
	task, exists := tl.Tasks.Get("0")
	if !exists {
		t.Fatalf("task not found")
	}
	ToggleTaskStatus(task.(*Task))
	if task.(*Task).Done != true {
		t.Fatalf("task not toggled")
	}
}
