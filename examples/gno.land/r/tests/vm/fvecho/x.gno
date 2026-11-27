// Package fvecho runs its own writer after another realm handed it back.
package fvecho

import "gno.land/r/tests/vm/fvechoy"

var Owner string

func Echo(cur realm) {
	f := func(s string) { Owner = s } // made in this crossing call
	g := fvechoy.Echo(f)              // same call, but fvechoy's code handed it back
	g("echo")
}
