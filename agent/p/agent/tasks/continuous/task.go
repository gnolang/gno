package continuous

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/agent/p/agent/task"
	"gno.land/p/demo/avl"
)

type Task struct {
	id                  string
	Era                 int
	NextDue             time.Time
	Interval            time.Duration
	Aggregator          task.Aggregator
	Definition          task.Definition
	RespondentWhiteList *avl.Tree
	Respondents         *avl.Tree
	History             task.History
}

func (t *Task) Finish(origCaller string) {
	now := time.Now()
	if now.Before(t.NextDue) {
		panic("era can not be transitioned until " + t.NextDue.String())
	}

	if t.RespondentWhiteList != nil {
		if !t.RespondentWhiteList.Has(origCaller) {
			panic("caller not in whitelist")
		}
	}

	// Handle the task state transitions.
	t.Era++
	t.NextDue = now.Add(t.Interval)
	t.Respondents = avl.NewTree()

	resultValue := t.Aggregator.Aggregate()
	t.History.AddResult(task.Result{Value: resultValue, Time: now})

	return
}

func (t Task) GetResult() (result task.Result, hasResult bool) {
	if len(t.History.Results) == 0 {
		return
	}

	return t.History.Results[len(t.History.Results)-1], true
}

func (t Task) ID() string {
	return t.id
}

func (t Task) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)
	w.WriteString(
		`{"id":"` + t.id +
			`","type":"` + t.Definition.Type() +
			`","era":` + strconv.Itoa(t.Era) +
			`,"next_due":` + strconv.FormatInt(t.NextDue.Unix(), 10) +
			`,"interval":` + strconv.FormatInt(int64(t.Interval/time.Second), 10) +
			`,"definition":`,
	)
	taskDefinitionBytes, err := t.Definition.MarshalJSON()
	if err != nil {
		return nil, err
	}

	w.Write(taskDefinitionBytes)
	w.WriteString("}")
	w.Flush()
	return buf.Bytes(), nil
}

func (t *Task) SubmitResult(origCaller, value string) {
	var era int
	valueParts := strings.SplitN(value, ",", 1)
	if len(valueParts) != 2 {
		panic("invalid result value; must be prefixed with era + ','")
	}

	era, err := strconv.Atoi(valueParts[0])
	if err != nil {
		panic(valueParts[0] + " is not a valid era")
	}

	value = valueParts[1]

	// Check that the era is for the next expected result.
	if era != t.Era {
		panic("expected era of " + strconv.Itoa(t.Era) + ", got " + strconv.Itoa(era))
	}

	// Check that the window to write has not ended.
	if time.Now().After(t.NextDue) {
		panic("era " + strconv.Itoa(t.Era) + " has ended; call Finish to transition to the next era")
	}

	if t.RespondentWhiteList != nil {
		if !t.RespondentWhiteList.Has(origCaller) {
			panic("caller not in whitelist")
		}
	}

	// Each agent can only respond once during each era.
	if t.Respondents.Has(origCaller) {
		panic("response already sent for this era")
	}

	t.Aggregator.AddValue(value)
	t.Respondents.Set(origCaller, struct{}{})
}
