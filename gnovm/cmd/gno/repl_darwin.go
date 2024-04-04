// go:build darwin
package main

import "syscall"

var (
	getTermios = syscall.TIOCGETA
	setTermios = syscall.TIOCSETA
)
