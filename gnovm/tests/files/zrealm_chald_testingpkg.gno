// PKGPATH: gno.land/r/tests/vm/chaldtesting
package chaldtesting

import (
	"testing"

	lib "gno.land/p/demo/tests/chaldlib"
)

var count int
var saved func()

func init() { saved = func() { count++ } }

func show(s string, err any) { println(s, count, err) }

// A stored function value is refused whether a library or the test-only
// stdlib "testing" calls it: only the runner's own call of a test is exempt.
// RunExampleTest recovers the refusal and prints it to stderr (above t03).
func main(cur realm) {
	show("t01 lib.Call(saved) control", try(func() { lib.Call(saved) }))
	show("t02 testing.AllocsPerRun2(saved)", try(func() { testing.AllocsPerRun2(1, saved) }))
	res := "not refused"
	if testing.RunExampleTest(saved) {
		res = "refused: read-only here, see the panic above"
	}
	show("t03 testing.RunExampleTest(saved)", res)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// t01 lib.Call(saved) control 0 gno.land/r/tests/vm/chaldtesting's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldtesting) package{})).count
// t02 testing.AllocsPerRun2(saved) 0 gno.land/r/tests/vm/chaldtesting's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldtesting) package{})).count
// panic: gno.land/r/tests/vm/chaldtesting's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/chaldtesting) package{})).count
// Stacktrace:
// exampleTestFunc<VPBlock(1,0)>()
//     gno.land/r/tests/vm/chaldtesting/zrealm_chald_testingpkg.gno:13
// ref(testing).RunExampleTest(saved<~VPBlock(4,1)>)
//     testing/testing.gno:389
// main<VPBlock(1,4)>(&(struct{("g1487taf6yryc4qp3pe2ef5zs0wdy7hjvqzl4hwx" .uverse.address),("gno.land/r/tests/vm/chaldtesting" string),(&(struct{("" .uverse.address),("" string),(undefined)} .uverse..grealm) *.uverse..grealm)} .uverse..grealm))
//     gno.land/r/tests/vm/chaldtesting/zrealm_chald_testingpkg.gno:24
//
// t03 testing.RunExampleTest(saved) 0 refused: read-only here, see the panic above
