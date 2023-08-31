package std

func Fn()

func FnRet() int

func FnParam(n int)

func FnParamRet(n int) int

func FnMachine()

func FnMachineRet() int

func FnMachineParam(n int)

func FnMachineParamRet(n int) int

func Ignored() int {
	// Ignored even if it has a matching go definition -
	// as this one has a body.
	return 1
}
