package todolist

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
)

type TodoList struct {
	Title string
	Tasks *avl.Tree
	Owner std.Address
}

type Task struct {
	Title string
	Done  bool
}

func NewTodoList(title string) *TodoList {
	return &TodoList{
		Title: title,
		Tasks: avl.NewTree(),
		Owner: std.GetOrigCaller(),
	}
}

func NewTask(title string) *Task {
	return &Task{
		Title: title,
		Done:  false,
	}
}

func (tl *TodoList) AddTask(id int, task *Task) {
	tl.Tasks.Set(strconv.Itoa(id), task)
}

func ToggleTaskStatus(task *Task) {
	task.Done = !task.Done
}

func (tl *TodoList) RemoveTask(taskId string) {
	tl.Tasks.Remove(taskId)
}

func (tl *TodoList) GetTasks() []*Task {
	tasks := make([]*Task, 0, tl.Tasks.Size())
	tl.Tasks.Iterate("", "", func(key string, value interface{}) bool {
		tasks = append(tasks, value.(*Task))
		return false
	})
	return tasks
}

func (tl *TodoList) GetTodolistOwner() std.Address {
	return tl.Owner
}

func (tl *TodoList) GetTodolistTitle() string {
	return tl.Title
}
