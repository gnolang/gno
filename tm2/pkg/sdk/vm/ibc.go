package vm

import "time"
import "github.com/gnolang/gno/tm2/pkg/sdk"

// simulate IBC callbacks
// TODO: implement interface, make it real

type IBC struct {
	VMKpr   *VMKeeper
	IBCChan chan MsgCall
}

func NewIBC() *IBC {
	return &IBC{
		IBCChan: make(chan MsgCall),
	}
}

func (i *IBC) SendPacket(msgCall MsgCall) {
	i.IBCChan <- msgCall
}

// this is called by IBC module
func (i *IBC) OnRecvPacket() {
	println("onRecvPacket")
	timeout := 3 * time.Second
	var mc MsgCall
	select {
	case mc = <-i.IBCChan:
	case <-time.After(timeout):
		panic("Timeout! Operation took too long.")
	}
	println("mc: ", mc.PkgPath, mc.Func, mc.Args[0])
	// dispatch msg
	r := i.VMKpr.dispatcher.HandleInnerMsgs(i.VMKpr.ctx, "xxx", []MsgCall{mc}, sdk.RunTxModeDeliver)
	println("result of call b is: ", string(r.Data))
}

func (i *IBC) OnAcknowledgementPacket() {

}
