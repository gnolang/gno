package ghverify

import (
	"bufio"
	"bytes"
)

type Task struct {
	gnoAddress   string
	githubHandle string
}

func NewTask(gnoAddress, githubHandle string) *Task {
	return &Task{
		gnoAddress:   gnoAddress,
		githubHandle: githubHandle,
	}
}

func (t *Task) MarshalToJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	w.Write(
		[]byte(`{"gnoAddress":"` + t.gnoAddress + `","githubHandle":"` + t.githubHandle + `"}`),
	)

	w.Flush()
	return buf.Bytes(), nil
}

func (t *Task) GnoAddress() string {
	return t.gnoAddress
}

func (t *Task) GithubHandle() string {
	return t.githubHandle
}
