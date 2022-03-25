package tests

import "std"

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
