package tests

import (
	"chain/runtime"

	psubtests "gno.land/p/demo/tests/subtests"
)

const World = "world"

func CurrentRealmPath() string {
	return runtime.CurrentRealm().PkgPath()
}

//----------------------------------------
// cross realm test vars

type TestRealmObject2 struct {
	Field string
}

func (o2 *TestRealmObject2) Modify() {
	o2.Field = "modified"
}

var (
	somevalue1 TestRealmObject2
	SomeValue2 TestRealmObject2
	SomeValue3 *TestRealmObject2
)

func init() {
	somevalue1 = TestRealmObject2{Field: "init"}
	SomeValue2 = TestRealmObject2{Field: "init"}
	SomeValue3 = &TestRealmObject2{Field: "init"}
}

func ModifyTestRealmObject2a() {
	somevalue1.Field = "modified"
}

func ModifyTestRealmObject2b() {
	SomeValue2.Field = "modified"
}

func ModifyTestRealmObject2c() {
	SomeValue3.Field = "modified"
}

func GetPreviousRealm() runtime.Realm {
	return runtime.PreviousRealm()
}

func GetPSubtestsPreviousRealm() runtime.Realm {
	return psubtests.GetPreviousRealm()
}

// Warning: unsafe pattern.
func Exec(fn func()) {
	fn()
}
