package crossrealm_slice

import "gno.land/p/demo/tests/p_crossrealm"

type Fooer interface{ Foo() }

var S []Fooer

func SetSlice(fs []Fooer) {
	S = fs
	for _, f := range fs {
		println(f)
	}
}

// -------------------------------------
type XYZ struct{ name string }

var s1 []XYZ
var s3 []XYZ

// var s4 []XYZ
var s4 = make([]XYZ, 1, 2)

// NOTE that if new list allocated, XYZ{"4"} will be copied, and it's a new object
// var s4 = make([]XYZ, 1)
var s5 []*XYZ
var s6 = make([]*XYZ, 2)

var s7 [2]XYZ

var s10 []int

func init() {
	s1 = append(s1, XYZ{"1"})
	s1 = append(s1, XYZ{"2"})

	s3 = append(s3, XYZ{"3"})
	//s4 = append(s4, XYZ{"4"})
	s4[0] = XYZ{"4"}

	s5 = append(s5, &XYZ{"5"})

	s6[0] = &XYZ{"6"}

	s7[0] = XYZ{"7"}
	s7[1] = XYZ{"7.1"} // should be real after this
	s10 = append(s10, 1)
}

func GetSlice() []XYZ {
	return s1[1:]
}

// -------------------------------------------------------
var e1 = XYZ{"1"}
var e2 = XYZ{"2"}

func GetSlice2() []XYZ {
	var s2 []XYZ
	s2 = append(s2, e1)
	s2 = append(s2, e2)
	return s2
}

// ------------------------------------------------------
func GetSlice3() []XYZ {
	s3 = append(s3, XYZ{"0"}) // this is real after function
	return s3
}

func GetSlice4(f func(s []XYZ)) {
	s4 = append(s4, XYZ{"0"}) // this is real after this function
	f(s4)                     // XYZ{"0"} is floating
}

func GetSlice5() []*XYZ { // TODO, unwrap
	return s5
}

func GetSlice6(f func(s []*XYZ)) {
	s6[1] = &XYZ{"6.1"}
	f(s6) // not real now
}

func GetSlice7(f func(s [2]XYZ)) {
	//s7[1] = XYZ{"7.1"}
	f(s7)
}

var s8 [1]p_crossrealm.Stringer

func GetSlice8(f p_crossrealm.Stringer) [1]p_crossrealm.Stringer {
	s8[0] = f
	return s8
}

func GetSlice9(f p_crossrealm.Stringer) [1]p_crossrealm.Stringer {
	var s9 [1]p_crossrealm.Stringer
	s9[0] = f
	return s9
}

// XXX, s10 under array is real, and owned by this realm.
func GetSlice10() []int {
	return s10
}

func GetSlice11(f func(s []int)) {
	// XXX, this is valid cuz the type of
	// array is not defined in current realm(or any other),
	// and the underlying array is not owned by this realm.
	s11 := []int{1}
	f(s11)
}
