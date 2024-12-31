package iface

// XXX, how about not interface
type Fooer interface{ Foo() }

var fooer Fooer

func SetFooer(f Fooer) Fooer {
	fooer = f
	return fooer
}

func CallFoo() { fooer.Foo() }

// container in external realm
type Container struct {
	name string
	f    Fooer
}

var container Container // attached

func SetContainer(f Fooer) {
	container.f = f // update
}

func SetContainer2(f Fooer) {
	ff := f          // associate to non-attached object
	SetContainer(ff) // attach container, while ff is not attached
}
