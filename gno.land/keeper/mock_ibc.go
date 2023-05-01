package vmk

import (
	vmh "github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"strconv"
)

type Packet struct {
	sequence int
	data     []byte
}

// Use channel to simulate IBC loop, the IBC/TAO layer
// TODO: saticefy a real IBCModule interface
type IBC struct {
	sendMsgQueue chan vmh.MsgCall // simulate call from IBC
	ackMsgQueue  chan *Packet
	vmk          *VMKeeper
	cbm          map[int]vmh.MsgCall // sequence -> callabck
}

func NewIBCModule(v *VMKeeper) *IBC {
	return &IBC{
		sendMsgQueue: make(chan vmh.MsgCall),
		ackMsgQueue:  make(chan *Packet),
		vmk:          v,
		cbm:          make(map[int]vmh.MsgCall),
	}
}

// simulate send out packet
// TODO: here the msgCall should be transcribed to shared types
// a call should be assgined with a unique sequence number, or use the IBC packet sequence number
func (i *IBC) SendPacket(msg vmh.MsgCall) {
	i.sendMsgQueue <- msg
	println("send packet done")
}

// b.gno
// XXX: need a portID and sequence to identify the initial call?
func (i *IBC) OnRecvPacket() {
	println("onRecvPacket")
	// timeout := 3 * time.Second
	for {
		select {
		// case msgCall := <-i.sendMsgQueue:

		// r := i.vmk.dispatcher.HandleInternalMsgs(i.vmk.ctx, []vmh.MsgCall{msgCall}, sdk.RunTxModeDeliver)
		// println("r.Data :", string(r.Data))

		// ack, handled by OnAck on the counterpart chain

		// i.ackMsgQueue <- &Packet{sequence: 1, data: r.Data}

		// case <-time.After(timeout):
		// 	panic("Timeout! IBC took too long.")
		}
	}
}

// a.gno
// call back to the initial caller
// needs a bind of sequence and portID and cb signature to identify the callback
// than do the callback
func (i *IBC) OnAcknowledgementPacket() {
	for {
		select {
		case ack := <-i.ackMsgQueue:
			println("ack, sequence, data ", strconv.Itoa(ack.sequence), string(ack.data))
			// bridge ack to vmKeeper
			i.vmk.ibcResponseQueue <- ack.data
		}
	}
}
