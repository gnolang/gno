//go:build unix

package supervisor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// LockFile is the file the supervisor locks inside the data directory.
const LockFile = "gnosupervisor.lock"

// Lock takes an exclusive, non-blocking lock on dataDir, so that two
// supervisors never run nodes on the same data directory. The kernel drops
// the lock when the process ends, however it ends. It does not protect
// against a second machine signing with the same key.
//
// The returned file is handed to every gnoland child as an extra file
// descriptor: the lock is held by the open file, so it stays taken until the
// supervisor and its child have both exited. A supervisor killed outright
// cannot release the data directory while its node is still shutting down.
func Lock(dataDir string) (*os.File, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(dataDir, LockFile)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, fmt.Errorf("%s is locked: another supervisor runs a node on %s", path, dataDir)
		}
		return nil, fmt.Errorf("lock %s: %w", path, err)
	}
	return f, nil
}

// lockPath takes an exclusive lock on path, creating it, and waits for it
// until ctx is done. Closing the file releases it.
func lockPath(ctx context.Context, path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return f, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			f.Close()
			return nil, fmt.Errorf("lock %s: %w", path, err)
		}
		select {
		case <-ctx.Done():
			f.Close()
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}

// checkPrivateDir refuses a directory that someone else could write into:
// the supervisor executes what it finds in its cache, so the cache must be
// as private as the binary it runs. It must be a real directory owned by
// this user, writable by no one else.
func checkPrivateDir(dir string) error {
	fi, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	switch {
	case !fi.IsDir():
		return fmt.Errorf("%s is not a directory", dir)
	case fi.Mode().Perm()&0o022 != 0:
		return fmt.Errorf("%s is writable by group or others (%v); the supervisor runs what it holds", dir, fi.Mode().Perm())
	case !ok || int(st.Uid) != os.Geteuid():
		return fmt.Errorf("%s is not owned by this user; the supervisor runs what it holds", dir)
	}
	return nil
}
