package main

func main() {}

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

var v3 = func() int {
	return func() int {
		v4 := 11
		return v4
	}()
}()

var v4 = func() int {
	return v3
}()

var v5 = func() int {
	v6 := 11
	return func() int {
		return v6
	}()
}()

var v6 = func() int {
	return v5
}()

var other = func() {
	if true {
		something := 2
		print(something) // 2
	} else {
		print(something) // a string, but single shared 'st' masks the outer/global reference.
	}
}
var something = "a string"

var other1 = func() {
	if true {
		something1 := 2
		print(something1) // 2
	}
	print(something1) // a string, but single shared 'st' masks the outer/global reference.
}
var something1 = "a string"
