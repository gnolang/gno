package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func Fn() {
	println("call Fn")
}

func FnRet() int {
	println("call FnRet")
	return 1
}

func FnParam(n int) {
	println("call FnParam", n)
}

func FnParamRet(n int) int {
	println("call FnParamRet", n)
	return 1
}

func FnMachine(m *gno.Machine) {
	println("call FnMachine")
}

func FnMachineRet(m *gno.Machine) int {
	println("call FnMachineRet")
	return 1
}

func FnMachineParam(m *gno.Machine, n int) {
	println("call FnMachineParam", n)
}

func FnMachineParamRet(m *gno.Machine, n int) int {
	println("call FnMachineParamRet", n)
	return 1
}

func Ignored() int {
	// Ignored even if it has a matching go definition -
	// as gno's has a body.
	return 1
}
