package os

import (
	"fmt"
	"os"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_writeStdout(m *gno.Machine, p []byte) (int, error) {
	return fmt.Fprint(os.Stdout, p)
}

func X_writeStderr(m *gno.Machine, p []byte) (int, error) {
	return fmt.Fprint(os.Stderr, p)
}
