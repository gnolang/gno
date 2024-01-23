package gh

import (
	"bufio"
	"bytes"
)

type verificationTask struct {
	gnoAddress   string
	githubHandle string
}

func NewVerificationTask(gnoAddress, githubHandle string) *verificationTask {
	return &verificationTask{
		gnoAddress:   gnoAddress,
		githubHandle: githubHandle,
	}
}

func (t *verificationTask) MarshalToJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	w.Write(
		[]byte(`{"gnoAddress":"` + t.gnoAddress + `","githubHandle":"` + t.githubHandle + `"}`),
	)

	w.Flush()
	return buf.Bytes(), nil
}

func (t *verificationTask) GnoAddress() string {
	return t.gnoAddress
}

func (t *verificationTask) GithubHandle() string {
	return t.githubHandle
}
