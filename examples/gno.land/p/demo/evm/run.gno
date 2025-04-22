package evm

import "fmt"

// for `gno run`
func main() {
	code := []byte{0x60, 0x01, 0x60, 0x02, 0x01, 0x60, 42, 0} // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
	buf, err := Execute(code)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(buf)
}
