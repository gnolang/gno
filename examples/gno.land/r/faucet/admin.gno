package faucet

import "std"

func AdminSetInPause(inPause bool)           { assertIsAdmin(); gInPause = inPause }
func AdminSetMessage(message string)         { assertIsAdmin(); gMessage = message }
func AdminSetPerTransferSend(send std.Coins) { assertIsAdmin(); gPerTransferSend = send }
func AdminSetAdminAddr(addr std.Address)     { assertIsAdmin(); gAdminAddr = addr }
func AdminSetGiverAddr(addr std.Address)     { assertIsAdmin(); gGiverAddr = addr }

func assertIsAdmin() bool {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	return caller == gAdminAddr
}
