//go:build !unix

package diff

import (
	"os/exec"
	"time"
)

func configureCommand(cmd *exec.Cmd) {
	cmd.WaitDelay = 2 * time.Second
}

func signaled(error) bool { return false }
