package crossrealm_array

type Foo struct{ name string }

var f = Foo{"foo"}

func GetArray() [2]Foo {
	return [2]Foo{Foo{"a"}, Foo{"b"}}
}

func GetArray2() [2]*Foo {
	return [2]*Foo{&Foo{"a"}, &Foo{"b"}}
}

func GetArray3() [1]*Foo {
	return [1]*Foo{&f}
}
