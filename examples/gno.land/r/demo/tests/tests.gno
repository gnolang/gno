package tests

import (
	"std"

	rsubtests "gno.land/r/demo/tests/subtests"
)

var counter int

func IncCounter() {
	counter++
}

func Counter() int {
	return counter
}

func CurrentRealmPath() string {
	return std.CurrentRealmPath()
}

var initOrigCaller = std.GetOrigCaller()

func InitOrigCaller() std.Address {
	return initOrigCaller
}

func AssertOriginCall() {
	std.AssertOriginCall()
}

func IsOriginCall() bool {
	return std.IsOriginCall()
}

//----------------------------------------
// Test structure to ensure cross-realm modification is prevented.

type TestRealmObject struct {
	Field string
}

func ModifyTestRealmObject(t *TestRealmObject) {
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
	gTestNode1 = &TestNode{Name: "first"}
	gTestNode2 = &TestNode{Name: "second", Child: &TestNode{Name: "second's child"}}
}

func ModTestNodes() {
	tmp := &TestNode{}
	tmp.Child = gTestNode2.Child
	gTestNode3 = tmp // set to new-real
	// gTestNode1 = tmp.Child // set back to original is-real
	gTestNode3 = nil // delete.
}

func PrintTestNodes() {
	println(gTestNode2.Child.Name)
}

func GetPrevRealm() std.Realm {
	return std.PrevRealm()
}

func GetRSubtestsPrevRealm() std.Realm {
	return rsubtests.GetPrevRealm()
}

func Exec(fn func()) {
	fn()
}
