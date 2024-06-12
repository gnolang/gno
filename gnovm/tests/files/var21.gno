package main

func main() {}

var myVar = func() { hello(myDep) }

var myDep string

var myVar1 = func() {
	a := myDep1
}

var myDep1 string

var myVar2 = func() {
	aaa := ""

	switch myDep {
	case aaa:
		println(myDep2)
	}
}

var myDep2 string

var myVar3 = func() {
	for _, c := range myDep3 {
		println(c)
	}
}

var myDep3 string

var v1 = func() int {
	v2 := 11
	return v2
}()

var v2 = func() int {
	return v1
}()
