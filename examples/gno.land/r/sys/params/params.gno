package params

import "std"

func Set(key string, value interface{}) {
	assertGovDao()
	std.SetParam(key, value)
}

func assertGovDao() {
	panic("not implemented")
}
