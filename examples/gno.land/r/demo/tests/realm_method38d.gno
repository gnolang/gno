package tests

var abs nat

func (n nat) Add() nat {
	return []Word{0}
}

func GetAbs() nat {
	switchrealm()

	abs = []Word{0}
	return abs
}

func AbsAdd() nat {
	switchrealm()

	rt := GetAbs().Add()
	return rt
}
