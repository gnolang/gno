package tests

import (
	"std"

	"gno.land/p/demo/nestedpkg"
	rsubtests "gno.land/r/demo/tests/subtests"
)

var counter int

func IncCounter() {
	crossing()

	counter++
}

func Counter() int {
	crossing()

	return counter
}

func CurrentRealmPath() string {
	crossing()

	return std.CurrentRealm().PkgPath()
}

var initOriginCaller = std.OriginCaller()

func InitOriginCaller() std.Address {
	crossing()

	return initOriginCaller
}

func CallAssertOriginCall() {
	crossing()

	std.AssertOriginCall()
}

func CallIsOriginCall() bool {
	crossing()

	return std.PreviousRealm().IsUser()
}

func CallSubtestsAssertOriginCall() {
	crossing()

	cross(rsubtests.CallAssertOriginCall)()
}

func CallSubtestsIsOriginCall() bool {
	crossing()

	return cross(rsubtests.CallIsOriginCall)()
}

//----------------------------------------
// Test structure to ensure cross-realm modification is prevented.

type TestRealmObject struct {
	Field string
}

var TestRealmObjectValue TestRealmObject

func ModifyTestRealmObject(t *TestRealmObject) {
	crossing()

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

func InitTestNodes() {
	crossing()

	gTestNode1 = &TestNode{Name: "first"}
	gTestNode2 = &TestNode{Name: "second", Child: &TestNode{Name: "second's child"}}
}

func ModTestNodes() {
	crossing()

	tmp := &TestNode{}
	tmp.Child = gTestNode2.Child
	gTestNode3 = tmp // set to new-real
	// gTestNode1 = tmp.Child // set back to original is-real
	gTestNode3 = nil // delete.
}

func PrintTestNodes() {
	println(gTestNode2.Child.Name)
}

func GetPreviousRealm() std.Realm {
	crossing()

	return std.PreviousRealm()
}

func GetRSubtestsPreviousRealm() std.Realm {
	crossing()

	return cross(rsubtests.GetPreviousRealm)()
}

func Exec(fn func()) {
	// no realm switching.
	fn()
}

func ExecSwitch(fn func()) {
	crossing()

	fn()
}

func IsCallerSubPath() bool {
	crossing()

	return nestedpkg.IsCallerSubPath()
}

func IsCallerParentPath() bool {
	crossing()

	return nestedpkg.IsCallerParentPath()
}

func HasCallerSameNamespace() bool {
	crossing()

	return nestedpkg.IsSameNamespace()
}
