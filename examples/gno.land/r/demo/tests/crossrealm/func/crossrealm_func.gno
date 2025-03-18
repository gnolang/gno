package crossrealm_func

type F func() bool

var f F

func SetCallback(ff F) {
	f = ff
}

func ExecuteCallback() {
	f()
}

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
type MyStruct2 struct{ name string }

func (sv *MyStruct) M() string { return "a" }

var mysv *MyStruct = &MyStruct{}

func (sv2 MyStruct2) M2() string { return "a2" }

var mysv2 = MyStruct2{name: "struct_val"}

func GetMethod() func() string {
	return mysv.M
}

func GetMethod2() func() string {
	return mysv2.M2
}
