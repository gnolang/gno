package core

import "gno.land/p/aib/ibc/types"

func SendPacket(cur realm, msg types.MsgSendPacket) (uint64 /*sequence*/, error) {
	// TODO
	return 0, nil
}

func RecvPacket(cur realm, msg types.MsgRecvPacket) (types.ResponseResultType, error) {
	// TODO
	return types.RESPONSE_FAILURE, nil
}

func Acknowledgement(cur realm, msg types.MsgAcknowledgement) (types.ResponseResultType, error) {
	// TODO
	return types.RESPONSE_FAILURE, nil
}

func Timeout(cur realm, msg types.MsgTimeout) (types.ResponseResultType, error) {
	// TODO
	return types.RESPONSE_FAILURE, nil
}
