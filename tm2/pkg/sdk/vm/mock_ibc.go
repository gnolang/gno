package vm

import (
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// Use channel to simulate IBC loop, the IBC/TAO layer
type IBC struct {
	VMKpr   *VMKeeper
	IBCChan chan MsgCall
}

func NewIBC() *IBC {
	return &IBC{
		IBCChan: make(chan MsgCall),
	}
}

// TODO: here the msgCall should be transcribed to shared types
func (i *IBC) SendPacket(msgCall MsgCall) {
	i.IBCChan <- msgCall
}

// callback on receive packet from IBC
// XXX: need a portID and sequence to identify the initial call?
func (i *IBC) OnRecvPacket(wg *sync.WaitGroup) {
	defer wg.Done()
	println("onRecvPacket")
	timeout := 3 * time.Second
	var mc MsgCall
	select {
	case mc = <-i.IBCChan:
	case <-time.After(timeout):
		panic("Timeout! Operation took too long.")
	}
	println("mc: ", mc.PkgPath, mc.Func, mc.Args[0])
	// TODO: just do, vmk.Call...
	// dispatch msg
	r := i.VMKpr.dispatcher.HandleInnerMsgs(i.VMKpr.ctx, "xxx", []MsgCall{mc}, sdk.RunTxModeDeliver)
	println("result of call b is: ", string(r.Data))
}

func (i *IBC) OnAcknowledgementPacket() {
	// call back to the initial caller
	// needs a bind of sequence and portID and cb signature to identify the callback
	// than do the callback
}
