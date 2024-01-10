package gh

import (
	"bufio"
	"bytes"
	"std"

	"gno.land/p/demo/agent/agent"
	ghTask "gno.land/p/demo/agent/github/verification/task"
	"gno.land/p/demo/agent/tasks/singular"
	"gno.land/p/demo/avl"
)

const (
	// TODO: set this to the address of the gnorkle agent that is authorized to submit verification tasks.
	authorizedAgentAddress string = ""
	verified                      = "OK"
)

var (
	handleToAddress = avl.NewTree()
	addressToHandle = avl.NewTree()
	gnorkleInstance *agent.Instance
)

func init() {
	gnorkleInstance = agent.Init(std.GetOrigCaller())
}

// This is a callback function that is invoked at the end of the GnorkleAgentSubmitAction execution.
// It only cares about submissions because we want to save the result of the verification and clean
// up the finished task.
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
	gnorkleInstance.RemoveTask(task.ID())

	// TODO: should it also clean up tasks that failed to verify?
}

// This is called by the gnorkle agent. The payload contains a function and optional data.
// In the case of this contract, it should only be called with payloads of:
//   - request (returns all verification tasks in the queue)
//   - submit,<id>,<data> (submits a verification result for the given task id where data is "OK" or not)
func GnorkleAgentSubmitAction(payload string) string {
	return gnorkleInstance.HandleRequest(payload, updateVerifiedGHData)
}

// RequestVerification will request that the gnorkle agent verify the given handle/address pair.
// A new tasks is created to be processed by the agent and verify there is a github repo of
// github.com/<handle>/gno-verification with a file "address.txt" that contains the given address.
func RequestVerification(handle, address string) {
	if address == "" {
		address = string(std.GetOrigCaller())
	}

	gnorkleInstance.AddTask(
		singular.NewTask(
			handle,
			ghTask.Definition{Handle: handle, Address: address},
			authorizedAgentAddress,
		),
	)
}

// Render will return all of the verified handle -> gno address pairs.
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
