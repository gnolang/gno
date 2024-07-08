package testutils

import (
	"io"
	"os"
	"strings"
)

// CaptureStdoutAndStderr temporarily pipes os.Stdout and os.Stderr into a buffer.
// Imported from https://github.com/moul/u/blob/master/io.go.
func CaptureStdoutAndStderr() func() (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return func() (string, error) { return "", err }
	}

	done := make(chan error, 1)

	oldErr := os.Stderr
	oldOut := os.Stdout
	os.Stderr = w
	os.Stdout = w

	var buf strings.Builder
	go func() {
		_, err := io.Copy(&buf, r)
		r.Close()
		done <- err
	}()

	closer := func() (string, error) {
		os.Stderr = oldErr
		os.Stdout = oldOut
		w.Close()
		err := <-done
		return buf.String(), err
	}
	return closer
}
