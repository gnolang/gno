// go:build linux
package main

import "syscall"

var (
	getTermios = syscall.TCGETS
	setTermios = syscall.TCSETS
)
