// Package fvforge backs the view-only foreign-value conversion probes.
package fvforge

import "gno.land/r/tests/vm/fvx"

type Local fvx.T

// Forge runs non-crossing: called from another realm it runs view-only.
func Forge() *fvx.T {
	s := &struct{ N int }{5}
	return (*fvx.T)(s)
}

// ForgeLocal goes through a local named type derived from fvx.T.
func ForgeLocal() *fvx.T {
	l := &Local{N: 6}
	return (*fvx.T)(l)
}
