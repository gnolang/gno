package gnoland

import (
	"os"
	"testing"
)

// skipIfDropped skips a test that reads an upstream-only workflow the internal
// fork drops (124be791e, 797f7ac7eb): syncs bring in upstream's tests of those
// workflows, but not the workflows.
func skipIfDropped(t *testing.T, rel string) {
	t.Helper()

	if _, err := os.Stat(repoPath(t, rel)); os.IsNotExist(err) {
		t.Skipf("%s is not in this tree: the internal fork drops it", rel)
	}
}
