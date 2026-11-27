//go:build unix

package gnoweb_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStaticFileAliasSkipsFIFO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "home.md")
	require.NoError(t, os.WriteFile(path, []byte("# Old"), 0o644))
	target, err := gnoweb.NewStaticFileAlias(path)
	require.NoError(t, err)

	require.NoError(t, syscall.Mkfifo(path+".new", 0o644))
	require.NoError(t, os.Rename(path+".new", path))
	done := make(chan gnoweb.AliasTarget, 1)
	go func() {
		target, err := gnoweb.ResolveAlias(target)
		assert.Error(t, err)
		done <- target
	}()
	select {
	case target = <-done:
		assert.Equal(t, "# Old", target.Value)
	case <-time.After(5 * time.Second):
		t.Fatal("reading a FIFO blocked")
	}
}
