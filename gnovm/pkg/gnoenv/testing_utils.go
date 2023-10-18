package gnoenv

import (
	"os"
	"testing"
)

func tBackupEnvironement(t *testing.T, keys ...string) {
	t.Helper()

	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if ok {
			os.Unsetenv(key)
			t.Cleanup(func() { os.Setenv(key, value) })
		} else {
			t.Cleanup(func() { os.Unsetenv(key) })
		}
	}
}
