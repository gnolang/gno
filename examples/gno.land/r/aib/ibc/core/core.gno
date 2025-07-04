package core

import "gno.land/p/aib/ibc/types"

func SendPacket(msg types.MsgSendPacket) (uint64 /*sequence*/, error) {
	// TODO
	return 0, nil
}

func RecvPacket(msg types.MsgRecvPacket) (types.ResponseResultType, error) {
	// TODO
	return types.RESPONSE_FAILURE, nil
}

func Acknowledgement(msg types.MsgAcknowledgement) (types.ResponseResultType, error) {
	// TODO
	return types.RESPONSE_FAILURE, nil
}

func Timeout(msg types.MsgTimeout) (types.ResponseResultType, error) {
	// TODO
	return types.RESPONSE_FAILURE, nil
}
