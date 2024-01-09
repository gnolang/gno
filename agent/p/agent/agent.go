package agent

import (
	// TODO: replace with std

	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/std"
)

var (
	tasks        *avl.Tree
	adminAddress string
)

func HandleRequest(payload string) string {
	payloadParts := strings.SplitN(payload, ",", 1)
	if len(payloadParts) != 2 {
		panic("invalid agent payload")
	}

	switch function := payloadParts[0]; function {
	case "finish":
		FinishTask(payloadParts[1])
	case "request":
		return RequestTasks()
	case "submit":
		submitArgs := strings.SplitN(payloadParts[1], ",", 1)
		if len(submitArgs) != 2 {
			panic("invalid agent submission payload")
		}

		SubmitTaskValue(submitArgs[0], submitArgs[1])
	default:
		panic("unknown function " + function)
	}

	return ""
}

func FinishTask(id string) {
	task, ok := tasks.Get(id)
	if !ok {
		panic("task not found")
	}

	task.(Task).Finish(string(std.GetOrigCaller()))
}

func RequestTasks() string {
	buf := new(strings.Builder)
	buf.WriteString("[")
	first := true
	tasks.Iterate("", "", func(_ string, value interface{}) bool {
		if !first {
			buf.WriteString(",")
		}

		first = false
		task := value.(Task)
		taskBytes, err := task.MarshalJSON()
		if err != nil {
			panic(err)
		}

		buf.Write(taskBytes)
		return true
	})
	buf.WriteString("]")
	return buf.String()
}

func SubmitTaskValue(id, value string) {
	task, ok := tasks.Get(id)
	if !ok {
		panic("task not found")
	}

	task.(Task).SubmitResult(string(std.GetOrigCaller()), value)
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
