package todolist

import (
	"chain/runtime"
	"strconv"

	"gno.land/p/nt/avl/v0"
)

type TodoList struct {
	Title string
	Tasks *avl.Tree
	Owner address
}

type Task struct {
	Title string
	Done  bool
}

func NewTodoList(title string) *TodoList {
	return &TodoList{
		Title: title,
		Tasks: avl.NewTree(),
		Owner: runtime.OriginCaller(),
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
	tl.Tasks.Iterate("", "", func(key string, value any) bool {
		tasks = append(tasks, value.(*Task))
		return false
	})
	return tasks
}

func (tl *TodoList) GetTodolistOwner() address {
	return tl.Owner
}

func (tl *TodoList) GetTodolistTitle() string {
	return tl.Title
}
