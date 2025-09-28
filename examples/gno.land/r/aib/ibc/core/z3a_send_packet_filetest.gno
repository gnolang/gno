package main

import (
	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
)

// SendPacket assert that MsgSendPacket.ValidateBasic is invoked
func main() {
	packet := types.MsgSendPacket{}
	core.SendPacket(cross, packet)
}

// Error:
// validate client identifier: identifier cannot be blank
