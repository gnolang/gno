package tests

import (
	"std"

	"gno.land/p/demo/nestedpkg"
	rsubtests "gno.land/r/demo/tests/subtests"
)

var counter int

func IncCounter() {
	switchrealm()

	counter++
}

func Counter() int {
	switchrealm()

	return counter
}

func CurrentRealmPath() string {
	switchrealm()

	return std.CurrentRealm().PkgPath()
}

var initOriginCaller = std.OriginCaller()

func InitOriginCaller() std.Address {
	switchrealm()

	return initOriginCaller
}

func CallAssertOriginCall() {
	switchrealm()

	std.AssertOriginCall()
}

func CallIsOriginCall() bool {
	switchrealm()

	return std.PreviousRealm().IsUser()
}

func CallSubtestsAssertOriginCall() {
	switchrealm()

	rsubtests.CallAssertOriginCall()
}

func CallSubtestsIsOriginCall() bool {
	switchrealm()

	return rsubtests.CallIsOriginCall()
}

//----------------------------------------
// Test structure to ensure cross-realm modification is prevented.

type TestRealmObject struct {
	Field string
}

var TestRealmObjectValue TestRealmObject

func ModifyTestRealmObject(t *TestRealmObject) {
	switchrealm()

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
	switchrealm()

	gTestNode1 = &TestNode{Name: "first"}
	gTestNode2 = &TestNode{Name: "second", Child: &TestNode{Name: "second's child"}}
}

func ModTestNodes() {
	switchrealm()

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
	switchrealm()

	return std.PreviousRealm()
}

func GetRSubtestsPreviousRealm() std.Realm {
	switchrealm()

	return withswitch(rsubtests.GetPreviousRealm)()
}

func Exec(fn func()) {
	// no realm switching.
	fn()
}

func ExecSwitch(fn func()) {
	switchrealm()

	fn()
}

func IsCallerSubPath() bool {
	switchrealm()

	return nestedpkg.IsCallerSubPath()
}

func IsCallerParentPath() bool {
	switchrealm()

	return nestedpkg.IsCallerParentPath()
}

func HasCallerSameNamespace() bool {
	switchrealm()

	return nestedpkg.IsSameNamespace()
}
