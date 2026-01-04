// SPDX-License-Identifier: GNO License Version 1.0

// Package gnolang contains the implementation of the Gno Virtual Machine.
package gnolang

//go:generate -command stringer go run -modfile ../../../misc/devdeps/go.mod golang.org/x/tools/cmd/stringer
//go:generate stringer -type=Kind,Op,TransCtrl,TransField,VPType,Word -output string_methods.go .
