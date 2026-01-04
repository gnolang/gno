package main

type S struct {
	i  int
	ip *int
	a  [1]int
	ap *[1]int
	s  []int
}

func main() {
	intPtr := new(int)
	*intPtr = 9

	v1 := S{
		i:  4,
		ip: intPtr,
		a:  [1]int{5},
		ap: new([1]int),
		s:  []int{6},
	}

	// v1 and all members are addressable.
	println(&v1)
	println(&v1.i)
	println(*v1.ip)
	println(&v1.a[0])
	println(&v1.ap[0])
	println(&v1.s[0])
	println("")

	// Defining a struct as a pointer also makes a member addressable.
	println(&(&S{i: 4}).i)
	println("")

	// Print only the members that are addressable when S is not addressable.
	println(*S{ip: intPtr}.ip)
	println(&S{ap: new([1]int)}.ap[0])
	println(&S{s: []int{6}}.s[0])
	println("")

	// A struct value returned by a function is not addressable.
	// Only certain members are addressable.
	println(*getStruct().ip)
	println(&getStruct().ap[0])
	println(&getStruct().s[0])
	println("")

	// A struct pointer value returned by a function has all members addressable.
	println(&getStructPtr().i)
	println(*getStructPtr().ip)
	println(&getStructPtr().a[0])
	println(&getStructPtr().ap[0])
	println(&getStructPtr().s[0])
}

func getStruct() S {
	intPtr := new(int)
	*intPtr = 9

	return S{
		i:  4,
		ip: intPtr,
		a:  [1]int{5},
		ap: new([1]int),
		s:  []int{6},
	}
}

func getStructPtr() *S {
	intPtr := new(int)
	*intPtr = 9

	return &S{
		i:  4,
		ip: intPtr,
		a:  [1]int{5},
		ap: new([1]int),
		s:  []int{6},
	}
}

// Output:
// &(struct{(4 int),(&(9 int) *int),(array[(5 int)] [1]int),(&(array[(0 int)] [1]int) *[1]int),(slice[(6 int)] []int)} main.S)
// &(4 int)
// 9
// &(5 int)
// &(0 int)
// &(6 int)
//
// &(4 int)
//
// 9
// &(0 int)
// &(6 int)
//
// 9
// &(0 int)
// &(6 int)
//
// &(4 int)
// 9
// &(5 int)
// &(0 int)
// &(6 int)
