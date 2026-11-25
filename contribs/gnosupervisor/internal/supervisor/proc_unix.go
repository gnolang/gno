//go:build unix && !linux

package supervisor

import "syscall"

// childAttrs puts gnoland in its own process group (see proc_linux.go).
func childAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
