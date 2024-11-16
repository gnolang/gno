package crossrealm

type XYZ struct{ name string }

var s1 []XYZ
var s3 []XYZ
var s4 []XYZ
var s5 []*XYZ

func init() {
	s1 = append(s1, XYZ{"1"})
	s1 = append(s1, XYZ{"2"})

	s3 = append(s3, XYZ{"3"})
	s4 = append(s4, XYZ{"4"})

	s5 = append(s5, &XYZ{"5"})
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
