package singular

import (
	"bufio"
	"bytes"
	"time"

	"gno.land/p/demo/agent/aatask"
)

type Task struct {
	id         string
	definition aatask.Definition

	isInactive           bool
	result               *aatask.Result
	authorizedRespondent string
}

func NewTask(id string, definition aatask.Definition, authorizedRespondent string) *Task {
	return &Task{
		id:                   id,
		definition:           definition,
		authorizedRespondent: authorizedRespondent,
	}
}

func (t Task) Definition() aatask.Definition {
	return t.definition
}

func (t Task) Finish(_ string) {
	panic("singular tasks are implicitly finished when a result is submitted")
}

func (t Task) GetResult() (result aatask.Result, hasResult bool) {
	if t.result == nil {
		return
	}

	return *t.result, true
}

func (t Task) ID() string {
	return t.id
}

func (t Task) MarshalJSON() ([]byte, error) {
	if !t.isInactive {
		return nil, nil
	}

	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)
	w.WriteString(
		`{"id":"` + t.id +
			`","type":"` + t.definition.Type() +
			`","definition":`,
	)

	taskDefinitionBytes, err := t.definition.MarshalJSON()
	if err != nil {
		return nil, err
	}

	w.Write(taskDefinitionBytes)
	w.WriteString("}")
	w.Flush()
	return buf.Bytes(), nil
}

func (t *Task) SubmitResult(origCaller, value string) {
	if t.isInactive {
		panic("task is inactive")
	}

	if t.authorizedRespondent != origCaller {
		panic("caller not authorized to submit result")
	}

	t.result = &aatask.Result{
		Value: value,
		Time:  time.Now(),
	}

	t.isInactive = true
	return
}
