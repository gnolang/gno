package agent

import (
	// TODO: replace with std

	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/std"
)

var (
	tasks *avl.Tree

	// Not sure this is necessary.
	adminAddress string
)

const (
	FunctionFinish  = "finish"
	FunctionRequest = "request"
	FunctionSubmit  = "submit"
)

type PostRequestAction func(function string, task Task)

func HandleRequest(payload string, post PostRequestAction) string {
	payloadParts := strings.SplitN(payload, ",", 1)
	if len(payloadParts) != 2 {
		panic("invalid agent payload")
	}

	switch function := payloadParts[0]; function {
	case FunctionFinish:
		task := FinishTask(payloadParts[1])
		if post != nil {
			post(function, task)
		}
	case FunctionRequest:
		return RequestTasks()
	case FunctionSubmit:
		submitArgs := strings.SplitN(payloadParts[1], ",", 1)
		if len(submitArgs) != 2 {
			panic("invalid agent submission payload")
		}

		SubmitTaskValue(submitArgs[0], submitArgs[1])
		if post != nil {
			post(function, nil)
		}
	default:
		panic("unknown function " + function)
	}

	return ""
}

func FinishTask(id string) Task {
	task, ok := tasks.Get(id)
	if !ok {
		panic("task not found")
	}

	task.(Task).Finish(string(std.GetOrigCaller()))
	return task.(Task)
}

func RequestTasks() string {
	buf := new(strings.Builder)
	buf.WriteString("[")
	first := true
	tasks.Iterate("", "", func(_ string, value interface{}) bool {
		if !first {
			buf.WriteString(",")
		}

		task := value.(Task)
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

func SubmitTaskValue(id, value string) Task {
	task, ok := tasks.Get(id)
	if !ok {
		panic("task not found")
	}

	task.(Task).SubmitResult(string(std.GetOrigCaller()), value)
	return task.(Task)
}

func AddTask(task Task) {
	if tasks.Has(task.ID()) {
		panic("task id " + task.ID() + " already exists")
	}

	tasks.Set(task.ID(), task)
}

func RemoveTask(id string) {
	if _, removed := tasks.Remove(id); !removed {
		panic("task id " + id + " not found")
	}
}

func Init(admin std.Address, newTasks ...Task) {
	if tasks != nil {
		panic("already initialized")
	}

	adminAddress = string(admin)
	tasks = avl.NewTree()
	for _, task := range newTasks {
		if updated := tasks.Set(task.ID(), task); updated {
			panic("task id " + task.ID() + " already exists")
		}
	}
}
