package testrunner

import (
	"fmt"
	"os"
	"runtime/debug"
	"testing"
)

func RunTest(testFunc func(t *testing.T)) error {
	tests := []testing.InternalTest{{Name: "blah", F: testFunc}}

	m := testing.MainStart(testDeps{}, tests, nil, nil, nil)

	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if ok {
				fmt.Fprintf(os.Stderr, "panic recovered: %v\n", err)
				debug.PrintStack()
			}
		}
	}()

	exitCode := m.Run()
	if exitCode != 0 {
		return fmt.Errorf("test failed")
	}

	return nil
}
