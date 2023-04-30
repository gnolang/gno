package main

import (
	"os"
	"testing"
)

func createTmpDir(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "gno-mod-test")
	if err != nil {
		t.Error("Failed to create tmp dir for mod:", err)
	}

	cleanUpFn := func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Logf("Failed to clean up test %s: %v", t.Name(), err)
		}
	}

	return tmpDir, cleanUpFn
}
