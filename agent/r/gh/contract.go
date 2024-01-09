package gh

import (
	"bufio"
	"bytes"

	"github.com/gnolang/gno/agent/p/agent"
	ghTask "github.com/gnolang/gno/agent/p/agent/github/verification/task"
	"github.com/gnolang/gno/agent/p/agent/tasks/singular"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/std"
)

const (
	authorizedAgentAddress string = ""
	verified                      = "OK"
)

var (
	handleToAddress = avl.NewTree()
	addressToHandle = avl.NewTree()
)

func init() {
	agent.Init(std.GetOrigCaller())
}

func updateVerifiedGHData(function string, task agent.Task) {
	if function != agent.FunctionSubmit {
		return
	}

	result, hasResult := task.GetResult()
	if !hasResult {
		return
	}

	if result.Value != verified {
		return
	}

	definition, ok := task.Definition().(ghTask.Definition)
	if !ok {
		panic("unexpected task definition of type " + definition.Type())
	}

	handleToAddress.Set(definition.Handle, definition.Address)
	addressToHandle.Set(definition.Address, definition.Handle)

	// It's been verified so clean it up.
	agent.RemoveTask(task.ID())
}

func OrkleAgentSubmitAction(payload string) string {
	return agent.HandleRequest(payload, updateVerifiedGHData)
}

func RequestVerification(handle, address string) {
	if address == "" {
		address = string(std.GetOrigCaller())
	}

	agent.AddTask(
		singular.NewTask(
			handle,
			ghTask.Definition{Handle: handle, Address: address},
			authorizedAgentAddress,
		),
	)
}

func Render(_ string) string {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)
	w.WriteString(`{"verified":{`)
	first := true
	handleToAddress.Iterate("", "", func(key string, value interface{}) bool {
		if !first {
			w.WriteString(",")
		}

		w.WriteString(`"` + key + `":"` + value.(string) + `"`)
		first = false
		return true
	})

	w.WriteString(`}}`)
	w.Flush()
	return buf.String()
}
