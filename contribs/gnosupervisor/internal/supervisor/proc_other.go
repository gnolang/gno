//go:build !unix

package supervisor

import "syscall"

func childAttrs() *syscall.SysProcAttr { return nil }
