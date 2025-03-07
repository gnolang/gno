package os

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_write(m *gnolang.Machine, p []byte, isStderr bool) int {
	if isStderr {
		if w, ok := m.Output.(interface{ StderrWrite(p []byte) (int, error) }); ok {
			n, _ := w.StderrWrite(p)
			return n
		}
	}
	n, _ := m.Output.Write(p)
	return n
}
