package os

type stdoutWriter struct{}
type stderrWriter struct{}

var Stdout = &stdoutWriter{}
var Stderr = &stderrWriter{}

func (w *stdoutWriter) Write(p []byte) (n int, err error) {
	return writeStdout(p)
}

func (w *stderrWriter) Write(p []byte) (n int, err error) {
	return writeStderr(p)
}

func Sleep(duration int64) {
	return sleep(duration)
}

// native bindings
func writeStderr(p []byte) (n int, err error)
func writeStdout(p []byte) (n int, err error)
func sleep(duration int64)
