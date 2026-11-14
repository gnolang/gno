// Package ifacevictim runs caller-supplied Runner values on its own object in
// every shape an interface method can be reached.
package ifacevictim

import (
	"bytes"
	"io"

	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/p/demo/tests/pfuncbind"
)

var gObj = &launderpkg.Immutable{Field: "victim-orig"}

func Read() string { return gObj.Field }

type Runner interface{ Run(*launderpkg.Immutable) }

type W struct{ Runner }

type Sub interface{ Runner }

func UseMV(r Runner)                    { f := r.Run; f(gObj) }
func UseEmbed(r Runner)                 { w := W{r}; w.Run(gObj) }
func UseEmbedPtr(r Runner)              { w := &W{r}; w.Run(gObj) }
func UseDefer(r Runner)                 { defer r.Run(gObj) }
func UseAssert(x any)                   { x.(Runner).Run(gObj) }
func UseSwitch(x any) {
	switch v := x.(type) {
	case Runner:
		v.Run(gObj)
	}
}
func UseMap(m map[string]Runner) { m["a"].Run(gObj) }
func UseSub(r Sub)               { r.Run(gObj) }
func UseMVArg(r Runner)          { call(r.Run) }
func call(f func(*launderpkg.Immutable)) { f(gObj) }

var gRunner Runner

func StoreX(cur realm, r Runner) { gRunner = r }
func UseStored()                 { gRunner.Run(gObj) }

// legit: stdlib wrapper writer over realm-owned buffer
var gBuf = &bytes.Buffer{}

func WriteMulti() string {
	var w io.Writer = io.MultiWriter(gBuf)
	w.Write([]byte("hi"))
	return gBuf.String()
}

func UseMVApply(r Runner) { pfuncbind.Apply(r.Run, gObj) }

var gF func(*launderpkg.Immutable)

func StoreMV(r Runner) { gF = r.Run }
func UseMVStored()     { gF(gObj) }
