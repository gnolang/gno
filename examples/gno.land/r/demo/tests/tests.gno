package tests

import (
	"std"

	"gno.land/r/demo/tests/subtests"
)

func CurrentRealmPath() string {
	return std.CurrentRealmPath()
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
// Test helpers to test a particualr realm bug.

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

func GetCaller() std.Address {
	return std.GetCaller()
}

func GetSubtestsCaller() std.Address {
	return subtests.GetCaller()
}

func ExecFromTest() {
	Exec(func() {
		println("ExecFromTest", std.GetCaller())
	})
}

func Exec(fn func()) {
	fn()
}

func PrintCallers() {
	println("p/TEST   : ORIGCALL:", std.GetOrigCaller())
	println("p/TEST   : PKG ADDR:", std.GetOrigPkgAddr())
	println("p/TEST   : CALLER  :", std.GetCaller())
}
