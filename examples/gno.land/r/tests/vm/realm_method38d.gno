package vm

var abs nat

func (n nat) Add() nat {
	return []Word{0}
}

func GetAbs(cur realm) nat {
	abs = []Word{0}
	return abs
}

func AbsAdd(cur realm) nat {
	rt := GetAbs(cur).Add()
	return rt
}
