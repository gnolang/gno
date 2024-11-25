package crossrealm

var ff = func() string { return "a" } // parent object already escaped

// XXX, this is special case, for ff, it's closure is referenced,
// so need to check all embedded items attached?
// see zrealm_crossrealm38.gno
func GetFunc() func() string {
	return ff
}

func GetFunc2() func() string {
	return func() string {
		return "b"
	}
}

type MyStruct struct{}

func (sv *MyStruct) M() string { return "a" }

var mysv *MyStruct = &MyStruct{}

//var f_local func() string
//
//func init() {
//	f_local = mysv.M
//}

func GetMethod() func() string {
	return mysv.M
}
