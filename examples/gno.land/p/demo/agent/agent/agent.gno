package agent

import (
	// TODO: replace with std

	"std"
	"strings"

	"gno.land/p/demo/avl"
)

type Instance struct {
	tasks *avl.Tree

	// Not sure this is necessary.
	adminAddress string
}

const (
	FunctionFinish  = "finish"
	FunctionRequest = "request"
	FunctionSubmit  = "submit"
)

type PostRequestAction func(function string, task Task)

func (i *Instance) HandleRequest(payload string, post PostRequestAction) string {
	payloadParts := strings.SplitN(payload, ",", 2)
	if len(payloadParts) == 0 {
		panic("invalid empty agent payload")
	}

	switch function := payloadParts[0]; function {
	case FunctionRequest:
		return i.RequestTasks()
	case FunctionFinish:
		if len(payloadParts) != 2 {
			panic("invalid agent finish payload")
		}

		task := i.FinishTask(payloadParts[1])
		if post != nil {
			post(function, task)
		}
	case FunctionSubmit:
		if len(payloadParts) != 2 {
			panic("invalid agent submission payload")
		}

		submitArgs := strings.SplitN(payloadParts[1], ",", 2)
		if len(submitArgs) != 2 {
			panic("invalid agent submission args")
		}

		task := i.SubmitTaskValue(submitArgs[0], submitArgs[1])
		if post != nil {
			post(function, task)
		}
	default:
		panic("unknown agent payload function " + function)
	}

	return ""
}

func (i *Instance) FinishTask(id string) Task {
	task, ok := i.tasks.Get(id)
	if !ok {
		panic("task not found")
	}

	task.(Task).Finish(string(std.GetOrigCaller()))
	return task.(Task)
}

func (i *Instance) RequestTasks() string {
	buf := new(strings.Builder)
	buf.WriteString("[")
	first := true

	i.tasks.Iterate("", "", func(_ string, value interface{}) bool {
		if !first {
			buf.WriteString(",")
		}

		task, ok := value.(Task)
		if !ok {
			panic("invalid task type")
		}

		taskBytes, err := task.MarshalJSON()
		if err != nil {
			panic(err)
		}

		// Guard against any tasks that shouldn't be returned; maybe they are not active because they have
		// already been completed.
		if len(taskBytes) == 0 {
			return true
		}

		first = false
		buf.Write(taskBytes)
		return true
	})
	buf.WriteString("]")
	return buf.String()
}

func (i *Instance) SubmitTaskValue(id, value string) Task {
	task, ok := i.tasks.Get(id)
	if !ok {
		panic("task not found")
	}

	task.(Task).SubmitResult(string(std.GetOrigCaller()), value)
	return task.(Task)
}

func (i *Instance) AddTask(task Task) {
	if i.tasks.Has(task.ID()) {
		panic("task id " + task.ID() + " already exists")
	}

	i.tasks.Set(task.ID(), task)
}

func (i *Instance) RemoveTask(id string) {
	if _, removed := i.tasks.Remove(id); !removed {
		panic("task id " + id + " not found")
	}
}

func Init(admin std.Address, newTasks ...Task) *Instance {
	i := &Instance{
		adminAddress: string(admin),
	}

	tasks := avl.NewTree()
	for _, task := range newTasks {
		if updated := tasks.Set(task.ID(), task); updated {
			panic("task id " + task.ID() + " already exists")
		}
	}

	i.tasks = tasks
	return i
}
