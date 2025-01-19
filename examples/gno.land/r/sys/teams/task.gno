package teams

type TaskFunc func(t *Team) Cmd

type Task interface {
	call(t *Team) Cmd
}

type task struct {
	actionFunc TaskFunc
}

func (a task) call(t *Team) Cmd {
	return a.actionFunc(t)
}

func CreateTaskCmd(cmd ...Cmd) Task {
	switch len(cmd) {
	case 0:
		return nil
	case 1:
		return CreateTask(func(_ *Team) Cmd {
			return cmd[0]
		})
	default:
	}

	fns := make([]TaskFunc, len(cmd))
	for i, m := range cmd {
		fns[i] = func(_ *Team) Cmd {
			return m
		}
	}
	return CreateTask(fns...)
}

func CreateTask(fn ...TaskFunc) Task {
	switch len(fn) {
	case 0:
		return nil
	case 1:
		return &task{actionFunc: fn[0]}
	default:
	}

	actions := make([]Task, len(fn))
	for i, f := range fn {
		actions[i] = &task{actionFunc: f}
	}
	return ChainTasks(actions...)
}

func ChainTasks(actions ...Task) Task {
	switch len(actions) {
	case 0:
		return nil
	case 1:
		return actions[0]
	default:
	}

	return CreateTask(func(t *Team) Cmd {
		cmds := make([]Cmd, len(actions))
		for i, action := range actions {
			cmds[i] = action.call(t)
		}
		return cmds
	})
}
