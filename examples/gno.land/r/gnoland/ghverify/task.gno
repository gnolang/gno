package ghverify

import (
	"bufio"
	"bytes"
)

type verificationTask struct {
	gnoAddress   string
	githubHandle string
}

// MarshalJSON marshals the task contents to JSON.
func (t *verificationTask) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	w.Write(
		[]byte(`{"gno_address":"` + t.gnoAddress + `","github_handle":"` + t.githubHandle + `"}`),
	)

	w.Flush()
	return buf.Bytes(), nil
}
