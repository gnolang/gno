//go:build linux

package supervisor

import "syscall"

// childAttrs puts gnoland in its own process group, so that a terminal's
// Ctrl-C reaches only the supervisor, which forwards it once it has decided
// what the stop means. It also asks the kernel to send SIGTERM to gnoland if
// the supervisor dies without forwarding a signal (SIGKILL, OOM), so that no
// orphan node keeps signing while a restarted supervisor starts another one.
func childAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGTERM}
}
