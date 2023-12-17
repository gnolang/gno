package task

const panicTaskIndex string = "task index out of bounds"

type Registry struct {
	taskHistory []*History
	tasks       []*Instance
}

func (r *Registry) Register(id int, task *Instance) {
	r.tasks = append(r.tasks, task)
	r.taskHistory = append(r.taskHistory, &History{})
}

func (r *Registry) Task(id int) *Instance {
	if id < 0 || id > len(r.tasks)-1 {
		panic(panicTaskIndex)
	}

	return r.tasks[id]
}

func (r *Registry) TaskHistory(id int) *History {
	if id < 0 || id > len(r.taskHistory)-1 {
		panic(panicTaskIndex)
	}
	return r.taskHistory[id]
}
