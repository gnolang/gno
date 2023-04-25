package vmk

import (
	vmi "github.com/gnolang/gno/tm2/pkg/sdk/vm"
)

var internelMsgQueue chan vmi.MsgCall // channel between IBC and VM

func init() {
	internelMsgQueue = make(chan vmi.MsgCall)
}

func getIBCQueue() <-chan vmi.MsgCall {
	return internelMsgQueue
}

// Use channel to simulate IBC loop, the IBC/TAO layer
// TODO: saticefy a real IBCModule interface
type IBC struct {
	IBCChan chan vmi.MsgCall // simulate call from IBC
	vmk     vmi.VMKeeperI
}

func NewIBCModule(v vmi.VMKeeperI) *IBC {
	return &IBC{
		IBCChan: make(chan vmi.MsgCall),
		vmk:     v,
	}
}

// TODO: here the msgCall should be transcribed to shared types
func (i *IBC) SendPacket(msgCall vmi.MsgCall) {
	println("send packet")
	i.IBCChan <- msgCall
	println("send packet done")
}

// callback on receive packet from IBC
// XXX: need a portID and sequence to identify the initial call?
func (i *IBC) OnRecvPacket() {
	println("onRecvPacket")
	// timeout := 3 * time.Second
	var mc vmi.MsgCall
	for {
		select {
		case mc = <-i.IBCChan:
			println("mc: ", mc.PkgPath, mc.Func, mc.Args[0])
			i.vmk.DispatchIBCMsg(mc)

			// case <-time.After(timeout):
			// 	panic("Timeout! IBC took too long.")
		}
	}
}

func (i *IBC) OnAcknowledgementPacket() {
	// call back to the initial caller
	// needs a bind of sequence and portID and cb signature to identify the callback
	// than do the callback
}
