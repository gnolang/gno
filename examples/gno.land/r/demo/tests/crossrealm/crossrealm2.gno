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
	f Fooer
}

var container Container // TODO, check does this attach or not

func AttachContainer() {
	container = Container{}
}
func SetContainer(f Fooer) {
	container.f = f
}
