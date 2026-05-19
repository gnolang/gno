package tests

import (
	"chain/banker"
	"chain/runtime"

	"gno.land/p/demo/nestedpkg"
	rsubtests "gno.land/r/tests/vm/subtests"
)

var counter int

func IncCounter(cur realm) {
	counter++
}

func Counter(cur realm) int {
	return counter
}

func CurrentRealmPath(cur realm) string {
	return runtime.CurrentRealm().PkgPath()
}

var initOriginCaller = runtime.OriginCaller()

func InitOriginCaller(cur realm) address {
	return initOriginCaller
}

func CallAssertOriginCall(cur realm) {
	runtime.AssertOriginCall()
}

func CallIsOriginCall(cur realm) bool {
	// XXX: consider return !runtime.PreviousRealm().IsCode()
	return runtime.PreviousRealm().IsUser()
}

func CallSubtestsAssertOriginCall(cur realm) {
	rsubtests.CallAssertOriginCall(cross2(cur))
}

func CallSubtestsIsOriginCall(cur realm) bool {
	return rsubtests.CallIsOriginCall(cross2(cur))
}

//----------------------------------------
// Test structure to ensure cross-realm modification is prevented.

type TestRealmObject struct {
	Field string
}

var TestRealmObjectValue TestRealmObject

func ModifyTestRealmObject(cur realm, t *TestRealmObject) {
	t.Field += "_modified"
}

func (t *TestRealmObject) Modify() {
	t.Field += "_modified"
}

//----------------------------------------
// Test helpers to test a particular realm bug.

type TestNode struct {
	Name  string
	Child *TestNode
}

var (
	gTestNode1 *TestNode
	gTestNode2 *TestNode
	gTestNode3 *TestNode
)

func InitTestNodes(cur realm) {
	gTestNode1 = &TestNode{Name: "first"}
	gTestNode2 = &TestNode{Name: "second", Child: &TestNode{Name: "second's child"}}
}

func ModTestNodes(cur realm) {
	tmp := &TestNode{}
	tmp.Child = gTestNode2.Child
	gTestNode3 = tmp // set to new-real
	// gTestNode1 = tmp.Child // set back to original is-real
	gTestNode3 = nil // delete.
}

func PrintTestNodes() {
	println(gTestNode2.Child.Name)
}

func GetPreviousRealm(cur realm) runtime.Realm {
	return runtime.PreviousRealm()
}

func GetRSubtestsPreviousRealm(cur realm) runtime.Realm {
	return rsubtests.GetPreviousRealm(cross2(cur))
}

func Exec(fn func()) {
	// no realm switching.
	fn()
}

// ExecRlm mirrors Exec but threads the caller's rlm into the callback
// so the callback can use `cross2(rlm)` instead of bare `cross`.
func ExecRlm(_ int, rlm realm, fn func(_ int, rlm realm)) {
	fn(0, rlm)
}

func ExecSwitch(cur realm, fn func()) {
	fn()
}

// ExecSwitchRlm is the rlm-threaded variant of ExecSwitch — crosses
// into this realm and passes the callee's cur to the callback so
// the callback can `cross2(rlm)` against the switched realm.
func ExecSwitchRlm(cur realm, fn func(_ int, rlm realm)) {
	fn(0, cur)
}

func IsCallerSubPath(cur realm) bool {
	return nestedpkg.IsCallerSubPath(0, cur)
}

func IsCallerParentPath(cur realm) bool {
	return nestedpkg.IsCallerParentPath(0, cur)
}

func HasCallerSameNamespace(cur realm) bool {
	return nestedpkg.IsSameNamespace(0, cur)
}

func BankerOriginSend(cur realm) string {
	return banker.OriginSend().String()
}

func RTestsOriginSend(cur realm) string {
	return rsubtests.BankerOriginSend(cross2(cur))
}
