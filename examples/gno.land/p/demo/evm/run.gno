package evm

import "fmt"

// for `gno run`
func main() {
	code := []byte{
		byte(PUSH1), 42, // data
		byte(PUSH1), 0, // offset
		byte(MSTORE),
		byte(PUSH1), 0,
		byte(MLOAD),
		byte(STOP),
	}
	st := newState()
	err := st.executeByteCodeWithGas(code, -1)
	if err != nil {
		fmt.Println("[ERROR]")
		fmt.Println(err)
		return
	}
	data := st.getStackData()
	fmt.Println("[")
	for _, b := range data {
		fmt.Println("\t", b.String())
	}
	fmt.Println("]")
	// buf, err := Execute(code)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(buf)
}
