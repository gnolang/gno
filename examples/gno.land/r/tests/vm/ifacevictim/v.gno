// Package ifacevictim runs caller-supplied Runner values on its own object in
// every shape an interface method can be reached.
package ifacevictim

import (
	"bytes"
	"io"

	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/p/demo/tests/pfuncbind"
	"gno.land/p/nt/avl/v0"
)

var gObj = &launderpkg.Immutable{Field: "victim-orig"}

func Read() string { return gObj.Field }

type Runner interface{ Run(*launderpkg.Immutable) }

type W struct{ Runner }

type Sub interface{ Runner }

func UseMV(cur realm, r Runner)       { f := r.Run; f(gObj) }
func UseEmbed(cur realm, r Runner)    { w := W{r}; w.Run(gObj) }
func UseEmbedPtr(cur realm, r Runner) { w := &W{r}; w.Run(gObj) }
func UseDefer(cur realm, r Runner)    { defer r.Run(gObj) }
func UseAssert(cur realm, x any)      { x.(Runner).Run(gObj) }
func UseSwitch(cur realm, x any) {
	switch v := x.(type) {
	case Runner:
		v.Run(gObj)
	}
}
func UseMap(cur realm, m map[string]Runner) { m["a"].Run(gObj) }
func UseSub(cur realm, r Sub)               { r.Run(gObj) }
func UseMVArg(cur realm, r Runner)          { call(r.Run) }
func call(f func(*launderpkg.Immutable))    { f(gObj) }

var gRunner Runner

func StoreX(cur realm, r Runner) { gRunner = r }
func UseStored(cur realm)        { gRunner.Run(gObj) }

// legit: stdlib wrapper writer over realm-owned buffer
var gBuf = &bytes.Buffer{}

func WriteMulti(cur realm) string {
	var w io.Writer = io.MultiWriter(gBuf)
	w.Write([]byte("hi"))
	return gBuf.String()
}

func UseMVApply(cur realm, r Runner) { pfuncbind.Apply(r.Run, gObj) }

var gF func(*launderpkg.Immutable)

func StoreMV(cur realm, r Runner) { gF = r.Run }
func UseMVStored(cur realm)       { gF(gObj) }

type HandlerGen interface {
	Handler() func(*launderpkg.Immutable)
}

func UseGen(cur realm, g HandlerGen)                    { h := g.Handler(); h(gObj) }
func ApplyOne(cur realm, f func(*launderpkg.Immutable)) { f(gObj) }

var gTree = avl.NewTree()

// Tree hands out the victim's tree; ApplySet runs a caller's setter on it.
func Tree() *avl.Tree                                  { return gTree }
func ApplySet(cur realm, f func(k string, v any) bool) { f("k", "pwnd") }
func TreeHas() bool                                    { return gTree.Has("k") }
