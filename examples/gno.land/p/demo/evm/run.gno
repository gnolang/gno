package evm

import "fmt"

// for `gno run`
func main() {
	code := []byte{0x5F, 0x60, 0x42, 0x00} // PUSH0, PUSH1 0x42, STOP
	buf, err := Execute(code)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(buf)
}
