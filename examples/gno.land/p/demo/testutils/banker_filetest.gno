package main 

import "std"
import "gno.land/p/demo/testutils"

func main(){
	mainAddr := std.TestDerivePkgAddr("main")
	addr1 := testutils.TestAddress("test1")

	banker := std.GetBanker(std.BankerTypeReadonly)

	balance := banker.GetCoins(mainAddr)
	println("balance of main: ", balance) // issued initially

	balance = banker.GetCoins(addr1)
	println("balance of test1 before send: ", balance)

	sent := std.GetOrigSend() // read from context
	println("OrigSend (read from ctx) before send is: ", sent)

	testutils.TestOrigSend(addr1, std.Coins{{"ugnot", 8}}, nil)

	balance = banker.GetCoins(addr1)
	println("balance of test1 after send: ", balance)

	sent = std.GetOrigSend() // read from context
	println("OrigSend (read from ctx) after send is: ", sent)
}

// Output:
// balance of main:  200000000ugnot
// balance of test1 before send:
// OrigSend (read from ctx) before send is:
// balance of test1 after send:  8ugnot
// OrigSend (read from ctx) after send is:  8ugnot