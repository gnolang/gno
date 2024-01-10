package task

import (
	"bufio"
	"bytes"
	"strconv"
)

const Type string = "random"

// Input in this range.
type Definition struct {
	RangeStart uint64
	RangeEnd   uint64
}

func (d Definition) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	w.WriteString(
		`{"start":` + strconv.FormatUint(d.RangeStart, 10) +
			`,"end":` + strconv.FormatUint(d.RangeEnd, 10) + `}`,
	)

	w.Flush()
	return buf.Bytes(), nil
}

func (d Definition) Type() string {
	return Type
}
