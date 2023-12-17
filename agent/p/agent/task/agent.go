package task

import (
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	// TODO: replace with std
	"gno.land/p/demo/std"
)

var (
	taskRegistry Registry
)

func TransitionEra(id int) (nextEra int, nextDueTime int64) {
	task := taskRegistry.Task(id)
	now := time.Now()
	if now.Before(task.NextDue) {
		panic("era can not be transitioned until " + task.NextDue.String())
	}

	// Handle the task state transitions.
	task.Era++
	task.NextDue = now.Add(task.Interval)
	task.Respondents = avl.NewTree()

	resultValue := task.Aggregator.Aggregate()
	taskRegistry.TaskHistory(id).AddResult(Result{Value: resultValue, Time: now})

	return task.Era, task.NextDue.Unix()
}

func SubmitTaskValue(id int, era int, value string) {
	task := taskRegistry.Task(id)

	// Check that the era is for the next expected result.
	if era != task.Era {
		panic("expected era of " + strconv.Itoa(task.Era) + ", got " + strconv.Itoa(era))
	}

	// Check that the window to write has not ended.
	if time.Now().After(task.NextDue) {
		panic("era " + strconv.Itoa(era) + " has ended; call TODO to transition to the next era")
	}

	// Each agent can only respond once during each era.
	var origCaller string
	if origCaller = string(std.GetOrigCaller()); task.Respondents.Has(origCaller) {
		panic("response already sent for this era")
	}

	task.Aggregator.AddValue(value)
	task.Respondents.Set(origCaller, struct{}{})
}
