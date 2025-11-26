package b // r/b

import "r/a"
import "r/c"

func Crossing(cur realm) {
	runtime.CurrentRealm() // => "r/b", DerivePkgAddr("r/b")
	runtime.PreviousRealm() // => "", "g1user1231321"
	
	a.NonCrossing() // switches to r/a, without r/a knowing about the caller
	a.A.Method()
	c.A.Method()
}

// -----------------------------------------------------------------------------

package a // r/a

import "chain/runtime"
import "p/users"

type MyType struct {
	A string
	B *users.Profile
	C readonly []byte
	D readonly *[123]byte
	E readonly map[string]int
}

func (m *MyType) Method() {
	// is setting m.A always legal?
	//  => "storage-realm": r/c
	m.A = "123"
	// is m.B.Name (always) legal?
	m.B.Name = "hellooo"
	m.C[0] = 123
	
	runtime.CurrentRealm() // => "r/a", DerivePkgAddr("r/a")
	runtime.PreviousRealm() // => "", ""
}

var A = &MyType{}

func NonCrossing() {
	runtime.CurrentRealm() // => "r/a", DerivePkgAddr("r/a")
	runtime.PreviousRealm() // => "", ""
}

// -----------------------------------------------------------------------------

package c // r/c

import "r/a"

var A = &a.MyType{
	B: &users.Profile{"Morgan", 123},
}

var G []byte = []byte("hello")

func GetString() readonly []byte {
	return G
}


// -----------------------------------------------------------------------------

package users // p/users

type Profile struct {
	Name string
	Age int
}

func (p *Profile) SetName(name string) { p.Name = name }