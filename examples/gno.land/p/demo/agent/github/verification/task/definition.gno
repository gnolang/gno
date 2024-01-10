package task

import (
	"bufio"
	"bytes"
)

const Type string = "gh-verification"

type Definition struct {
	Handle  string
	Address string
}

func (d Definition) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	w.WriteString(
		`{"handle":"` + d.Handle +
			`","address":"` + d.Address + `"}`,
	)

	w.Flush()
	return buf.Bytes(), nil
}

func (d Definition) Type() string {
	return Type
}
