package cbvictim

var Owner = "victim"

// SetOwner is exported and not crossing: callers outside must not write.
func SetOwner(s string) { Owner = s }

// Run is crossing and runs a function value the caller handed in.
func Run(cur realm, f func()) { f() }

// RunS is crossing and runs a function value the caller handed in.
func RunS(cur realm, f func(string)) { f("picked") }

// Getter returns a closure; made in read-only code it carries no authority.
func Getter() func(string) { return func(s string) { Owner = s } }

// Keep is a closure made in a gate: it writes for this realm.
var Keep func(string)

func Init(cur realm) { Keep = func(s string) { Owner = s } }

type T struct{ Name string }

var Box = &T{}

func (t *T) SetName(s string) { t.Name = s }

var hook func(string)

func SetHook(cur realm, f func(string)) { hook = f }
func RunHook(cur realm)                 { hook("stored") }

// Writers builds a closure writing Owner and one writing what it is given;
// entered without crossing, both carry no maker.
func Writers() (func(string), func([]string)) {
	return func(s string) { Owner = s }, func(b []string) { b[0] = "pwnd" }
}

// GetMV returns a method value bound while read-only: it has no binder.
func GetMV() func(string) { return Box.SetName }

type S struct{ F func(string) }

// GetSt returns a struct holding a method value bound while read-only.
func GetSt() *S { return &S{F: Box.SetName} }

// RunSt is crossing and runs the function a caller's struct holds.
func RunSt(cur realm, s *S) { s.F("struct") }

// GetDecl and FV hand out this realm's own declared function.
func GetDecl() func(string) { return SetOwner }

var FV = SetOwner
