package tests

import (
	"std"

	psubtests "gno.land/p/demo/tests/subtests"
	"gno.land/r/demo/tests"
	rtests "gno.land/r/demo/tests"
)

const World = "world"

// IncCounter demonstrates that it's possible to call a realm function from
// a package. So a package can potentially write into the store, by calling
// an other realm.
func IncCounter() {
	tests.IncCounter()
}

func CurrentRealmPath() string {
	return std.CurrentRealmPath()
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

func GetPrevRealm() std.Realm {
	return std.PrevRealm()
}

func GetPSubtestsPrevRealm() std.Realm {
	return psubtests.GetPrevRealm()
}

func GetRTestsGetPrevRealm() std.Realm {
	return rtests.GetPrevRealm()
}

// Warning: unsafe pattern.
func Exec(fn func()) {
	fn()
}
