package log

import (
	"io"
	"sync"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelError
)

type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	With(keyvals ...interface{}) Logger

	SetLevel(LogLevel)
}

//----------------------------------------

// NewSyncWriter returns a new writer that is safe for concurrent use by
// multiple goroutines. Writes to the returned writer are passed on to w. If
// another write is already in progress, the calling goroutine blocks until
// the writer is available.
func NewSyncWriter(w io.Writer) io.Writer {
	return &syncWriter{Writer: w}
}

// syncWriter synchronizes concurrent writes to an io.Writer.
type syncWriter struct {
	sync.Mutex
	io.Writer
}

// Write writes p to the underlying io.Writer. If another write is already in
// progress, the calling goroutine blocks until the syncWriter is available.
func (w *syncWriter) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	return w.Writer.Write(p)
}
