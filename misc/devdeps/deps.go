//go:build tools
// +build tools

package gno

// this file is used to allow "go mod" to be aware of some development dependencies.

import (
	// required by Makefile for flappy tests
	_ "moul.io/testman"
)
