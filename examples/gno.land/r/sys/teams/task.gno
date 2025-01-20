package teams

// TaskFunc defines a function type that takes a Team and returns a Cmd.
// It represents the executable logic that can be performed on a team.
type TaskFunc func(t *Team) Cmd

// Task represents a unit of work that can be executed to apply a command.
// Tasks encapsulate the execution logic of commands, ensuring that operations
// are performed with the correct permissions and in a specified order.
//
// Tasks are used to execute commands that modify the team's state, ensuring
// that these modifications adhere to the permissions set by the access control
// mechanisms and are executed in a controlled sequence.
type Task interface {
	call(t *Team) Cmd
}

type task struct {
	actionFunc TaskFunc
}

func (a task) call(t *Team) Cmd {
	return a.actionFunc(t)
}

// CreateTaskCmd creates a Task from one or more commands.
func CreateTaskCmd(cmd ...Cmd) Task {
	switch len(cmd) {
	case 0:
		return nil
	case 1:
		return CreateTask(func(_ *Team) Cmd {
			return cmd[0]
		})
	default:
		// Handle multiple commands
	}

	fns := make([]TaskFunc, len(cmd))
	for i, m := range cmd {
		fns[i] = func(_ *Team) Cmd {
			return m
		}
	}
	return CreateTask(fns...)
}

// CreateTask creates a Task from one or more TaskFuncs.
func CreateTask(fn ...TaskFunc) Task {
	switch len(fn) {
	case 0:
		return nil
	case 1:
		return &task{actionFunc: fn[0]}
	default:
		// Handle multiple functions
	}

	actions := make([]Task, len(fn))
	for i, f := range fn {
		actions[i] = &task{actionFunc: f}
	}
	return ChainTasks(actions...)
}

// ChainTasks creates a single Task that executes a series of tasks in sequence.
// It combines multiple tasks into one.
func ChainTasks(actions ...Task) Task {
	switch len(actions) {
	case 0:
		return nil
	case 1:
		return actions[0]
	default:
		// Handle chaining of multiple tasks
	}

	return CreateTask(func(t *Team) Cmd {
		cmds := make([]Cmd, len(actions))
		for i, action := range actions {
			cmds[i] = action.call(t)
		}
		return Cmds(cmds)
	})
}
