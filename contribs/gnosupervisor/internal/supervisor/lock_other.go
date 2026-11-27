//go:build !unix

package supervisor

import (
	"context"
	"errors"
	"os"
)

// LockFile is the file the supervisor locks inside the data directory.
const LockFile = "gnosupervisor.lock"

// Lock is not implemented outside unix: gnoland release images are linux only.
func Lock(string) (*os.File, error) {
	return nil, errors.New("the supervisor runs on unix systems only")
}

func lockPath(context.Context, string) (*os.File, error) {
	return nil, errors.New("the supervisor runs on unix systems only")
}

func checkPrivateDir(string) error {
	return errors.New("the supervisor runs on unix systems only")
}
