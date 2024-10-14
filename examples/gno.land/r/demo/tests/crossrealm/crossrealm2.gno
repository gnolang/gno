package crossrealm

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

var container Container // XXX, not attached here

func AttachContainer() {
	container = Container{name: "external_container"} // attach after call
}

func SetContainer(f Fooer) {
	container.f = f
}

func SetContainer2(f Fooer) {
	ff := f
}
