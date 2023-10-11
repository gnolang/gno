package gnoenv

import (
	"os"
	"testing"
)

func tBackupEnvironement(t *testing.T, keys ...string) {
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		if ok {
			t.Cleanup(func() { os.Setenv(key, value) })
		} else {
			t.Cleanup(func() { os.Unsetenv(key) })
		}
		os.Unsetenv(key)
	}
}
