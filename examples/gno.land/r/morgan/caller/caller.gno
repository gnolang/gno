package caller

var clos func()

func StoreClosure(cur realm, fn func()) {
	clos = fn
}

func CallClosure(cur realm) {
	clos()
}
