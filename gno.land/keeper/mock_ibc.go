package vmk

import (
	vmh "github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"strconv"
	"time"
)

// Use channel to simulate IBC loop, the IBC/TAO layer
// TODO: a real ibc module, satisfy IBC interfaces, bindPort, etc

type AckPacket struct {
	sequence int
	data     []byte
}

type IBCChannelKeeper struct {
	vmk          *VMKeeper
	sendMsgQueue chan vmh.MsgCall // bridge between ICB and VM keeper
	ackMsgQueue  chan *AckPacket  // only for inter-loop simulation
	// cbm          map[int]vmh.MsgCall // sequence -> callabck
}

func NewIBCChannelKeeper(v *VMKeeper) *IBCChannelKeeper {
	return &IBCChannelKeeper{
		sendMsgQueue: make(chan vmh.MsgCall),
		ackMsgQueue:  make(chan *AckPacket),
		vmk:          v,
		// cbm:          make(map[int]vmh.MsgCall),
	}
}

// simulate send out packet through IBC
// TODO: for Heterogeneous scenario, msgCall should be transcribed to shared types.
// a call should be assgined with an unique sequence number, or use the IBC packet sequence number
func (i *IBCChannelKeeper) SendPacket(msg vmh.MsgCall) {
	println("send packet")
	i.sendMsgQueue <- msg
	println("send packet done")
}

// callee side
// XXX: need a portID and sequence to identify the initial call?
func (i *IBCChannelKeeper) OnRecvPacket() {
	println("onRecvPacket")
	timeout := 3 * time.Second
	for {
		select {
		case msgCall := <-i.sendMsgQueue:
			r, err := i.vmk.Call(i.vmk.ctx, msgCall)
			if err != nil {
				panic(err.Error())
			}
			println("result :", string(r))

			// ack, handled by OnAck on the counterpart chain
			i.ackMsgQueue <- &AckPacket{sequence: 1, data: []byte(r)}

		case <-time.After(timeout):
			panic("Timeout! IBC took too long.")
		}
	}
}

// caller
// handle ack, callback to caller contract
func (i *IBCChannelKeeper) OnAcknowledgementPacket() {
	for {
		select {
		case ack := <-i.ackMsgQueue:
			println("ack, sequence, data: ", strconv.Itoa(ack.sequence), string(ack.data))
			// callback
			i.vmk.ibcResponseQueue <- string(ack.data)
		}
	}
}
