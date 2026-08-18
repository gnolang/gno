// PKGPATH: gno.land/r/decltypeiface
package decltypeiface

// Canary axis (identity). Companion to decltype.gno, which
// guards the positional axis via concrete dispatch.
//
// Gno code cannot craft a violation (DeclaredType is a Go struct, not
// reachable from Gno), so this test cannot *produce* drift; it notices
// drift introduced by a Go-side VM bug. Held.Say() is an interface call:
// at runtime the VM walks T.Methods by Name (see values.go VPInterface
// branch), not by a preprocess-bound index, and resolves the FuncValue
// from the current slice on every call. This catches identity
// mutations — Say removed, renamed, or its FuncValue replaced with a
// body that returns something else — which surface as a panic ("method
// Say not found") or a wrong printed string.

type Sayer interface {
	Say() string
}

type T struct {
	Msg string
}

func (t T) Say() string { return t.Msg }

// Compile-time: T must satisfy Sayer. Forces the VM to compute T's
// method set against Sayer's interface requirements during preprocess,
// which is the point the method set becomes final.
var _ Sayer = T{}

var Held Sayer = T{Msg: "hi"}

func main(cur realm) {
	println(Held.Say())
}

// Output:
// hi

// Types:
// Sayer[gno.land/r/decltypeiface.Sayer]=
//     {
//         "ID": "gno.land/r/decltypeiface.Sayer"
//     }
// T[gno.land/r/decltypeiface.T]=
//     {
//         "ID": "gno.land/r/decltypeiface.T"
//     }
