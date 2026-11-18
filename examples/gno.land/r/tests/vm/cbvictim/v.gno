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
