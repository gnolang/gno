package main 

import "std"
import "gno.land/p/demo/testutils"


func deposit(){
	sent := std.GetOrigSend() // read from context
	println("orig sent is: ", sent)
}

func main(){
	addr1 := testutils.TestAddress("test1")
	banker := std.GetBanker(std.BankerTypeReadonly)
	balance := banker.GetCoins(addr1)
	println("before send: ", balance)

	testutils.TestOrigSend(addr1, std.Coins{{"ugnot", 8}}, nil)

	balance = banker.GetCoins(addr1)
	println("after send: ", balance)

	deposit()
}

// Output:
// before send:
// after send:  8ugnot
// orig sent is:  8ugnot
