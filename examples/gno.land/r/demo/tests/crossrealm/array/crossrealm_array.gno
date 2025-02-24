package crossrealm_array

type Foo struct{ name string }

var f = Foo{"foo"}

var Arr = [2]Foo{Foo{"a"}, Foo{"b"}}

var Arr2 = [2]int{1, 2}

func GetArray5() [2]int {
	return Arr2
}

func GetArray4() [2]Foo {
	return Arr
}

func GetArray() [2]Foo {
	return [2]Foo{Foo{"a"}, Foo{"b"}}
}

func GetArray2() [2]*Foo {
	return [2]*Foo{&Foo{"a"}, &Foo{"b"}}
}

func GetArray3() [1]*Foo {
	return [1]*Foo{&f}
}
