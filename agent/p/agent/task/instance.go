package task

import (
	"bufio"
	"bytes"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
)

type Instance struct {
	ID          int
	Type        uint8
	Era         int
	NextDue     time.Time
	Interval    time.Duration
	Aggregator  Aggregator
	Definition  Definition
	Respondents *avl.Tree
}

func (i Instance) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)
	w.WriteString(
		`{"id":"` + strconv.Itoa(i.ID) +
			`","type":` + strconv.FormatUint(uint64(i.Type), 10) +
			`,"era":` + strconv.Itoa(i.Era) +
			`,"next_due":` + strconv.FormatInt(i.NextDue.Unix(), 10) +
			`,"interval":` + strconv.FormatInt(int64(i.Interval/time.Second), 10) +
			`,"definition":`,
	)
	taskDefinitionBytes, err := i.Definition.MarshalJSON()
	if err != nil {
		return nil, err
	}

	w.WriteString(string(taskDefinitionBytes) + "}")
	w.Flush()
	return buf.Bytes(), nil
}
