package tests

var abs nat

func (n nat) Add() nat {
	return []Word{0}
}

func GetAbs() nat {
	crossing()

	abs = []Word{0}
	return abs
}

func AbsAdd() nat {
	crossing()

	rt := GetAbs().Add()
	return rt
}
