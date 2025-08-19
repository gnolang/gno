package p_crossrealm

import (
	"errors"
	"std"
)

type Foo struct {
	num int
}

var F *Foo = &Foo{num: 0}

var Err = errors.New("Err 1")

func (f *Foo) Modify2() {
	f.num = 1
	println(f)
	Err = errors.New("Err 2")
	println(Err)
}

// -------------------------------
var FF = func() {
	AllowedList = append(AllowedList, 3)
}

//--------------------------------

var AllowedList = []int{1, 2}

func Modify() {
	AllowedList = append(AllowedList, 3)
	println(AllowedList)
}

type Stringer interface {
	String() string
}

type Container struct {
	A int
	B Stringer
}

func (c *Container) Touch() *Container {
	c.A += 1
	return c
}

func (c *Container) Print() {
	println("A:", c.A)
	if c.B == nil {
		println("B: undefined")
	} else {
		println("B:", c.B.String())
	}
}

func CurrentRealm() std.Realm {
	return std.CurrentRealm()
}
