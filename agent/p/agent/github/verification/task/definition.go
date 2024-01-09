package task

import (
	"bufio"
	"bytes"
)

const Type string = "gh-verification"

type Definition struct {
	Handle    string
	Signature string
}

func (d Definition) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	w.WriteString(
		`{"handle":"` + d.Handle +
			`","signature":"` + d.Signature + `"}`,
	)

	w.Flush()
	return buf.Bytes(), nil
}

func (d Definition) Type() string {
	return Type
}
