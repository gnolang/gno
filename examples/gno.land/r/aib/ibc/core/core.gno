package core

import "gno.land/p/aib/ibc/types"

func SendPacket(msg []byte /*types.MsgSendPacket*/) (uint64 /*sequence*/, error) {
	// TODO
	return 0, nil
}

func RecvPacket(msg []byte /*types.MsgRecvPacket*/) (types.ResponseResultType, error) {
	// TODO
	return types.RESPONSE_FAILURE, nil
}

func Acknowledgement(msg []byte /*types.MsgAcknowledgement*/) (types.ResponseResultType, error) {
	// TODO
	return types.RESPONSE_FAILURE, nil
}

func Timeout(msg []byte /*types.MsgTimeout*/) (types.ResponseResultType, error) {
	// TODO
	return types.RESPONSE_FAILURE, nil
}
