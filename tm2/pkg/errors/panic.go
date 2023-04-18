package errors

import "context"

var (
	fatalCtx      context.Context
	fatalCancelFn func()
)

func init() {
	fatalCtx, fatalCancelFn = context.WithCancel(context.Background())
}

// FatalContext returns a context cancelled by Fatal to quickly exit the program at the top-level.
func FatalContext() context.Context { return fatalCtx }

// Fatal is the highest level of errors, indicating behavior that should not be handled or recovered.
func Fatal(err error) error {
	fatalCancelFn()
	// additional logging could be done here.
	return err
}
