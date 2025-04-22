package evm

import "fmt"

// for `gno run`
func main() {
	code := []byte{0x60, 11, 0x60, 10, 0x01, 0x00} // PUSH0, PUSH1 0x42, STOP
	st := newState()
	err := st.executeByteCode(code)
	if err != nil {
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
