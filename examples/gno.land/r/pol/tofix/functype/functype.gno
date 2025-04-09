package functype

type mySuperFuncType func() // define a typed function

func funcAndInt() (mySuperFuncType, int) { // return this type and another crap
	return func(){}, 42
}

func defaultFunc() {
	var f mySuperFuncType
	f, i := funcAndInt() // try to get both: panic
	_ = i
	_ = f
}