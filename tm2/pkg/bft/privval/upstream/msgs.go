package upstream

// msgs.go: privval Message envelope wrap/unwrap helpers.
//
// Mirrors cometbft/privval/msgs.go. Concrete privval message types are
// wrapped in upstreampb.Message (a protoc-generated oneof) for wire I/O,
// and unwrapped at the receive side via type-switch on the Sum field.
//
// The protoc-generated Marshal/Unmarshal on upstreampb.Message handles the
// oneof wire-format correctly — exactly one Sum branch is encoded per
// message, matching upstream Tendermint v0.34's privval/types.proto.

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
)

// WrapMsg builds an upstreampb.Message containing the given concrete
// privval message. Panics if pb is not one of the recognized privval
// message types — this is a programming error, not a runtime input.
//
// Mirror of cometbft/privval/msgs.go::mustWrapMsg.
func WrapMsg(pb interface{}) *upstreampb.Message {
	msg := &upstreampb.Message{}
	switch m := pb.(type) {
	case *upstreampb.PubKeyRequest:
		msg.Sum = &upstreampb.Message_PubKeyRequest{PubKeyRequest: m}
	case *upstreampb.PubKeyResponse:
		msg.Sum = &upstreampb.Message_PubKeyResponse{PubKeyResponse: m}
	case *upstreampb.SignVoteRequest:
		msg.Sum = &upstreampb.Message_SignVoteRequest{SignVoteRequest: m}
	case *upstreampb.SignedVoteResponse:
		msg.Sum = &upstreampb.Message_SignedVoteResponse{SignedVoteResponse: m}
	case *upstreampb.SignProposalRequest:
		msg.Sum = &upstreampb.Message_SignProposalRequest{SignProposalRequest: m}
	case *upstreampb.SignedProposalResponse:
		msg.Sum = &upstreampb.Message_SignedProposalResponse{SignedProposalResponse: m}
	case *upstreampb.PingRequest:
		msg.Sum = &upstreampb.Message_PingRequest{PingRequest: m}
	case *upstreampb.PingResponse:
		msg.Sum = &upstreampb.Message_PingResponse{PingResponse: m}
	default:
		panic(fmt.Errorf("upstream.WrapMsg: unknown privval message type %T", pb))
	}
	return msg
}

// UnwrapMsg returns the concrete privval message inside an
// upstreampb.Message envelope. Returns an error if the envelope is empty
// or carries an unrecognized Sum variant — the wire was malformed or
// emitted by an incompatible peer.
func UnwrapMsg(msg *upstreampb.Message) (interface{}, error) {
	if msg == nil || msg.Sum == nil {
		return nil, fmt.Errorf("upstream.UnwrapMsg: empty envelope")
	}
	switch s := msg.Sum.(type) {
	case *upstreampb.Message_PubKeyRequest:
		return s.PubKeyRequest, nil
	case *upstreampb.Message_PubKeyResponse:
		return s.PubKeyResponse, nil
	case *upstreampb.Message_SignVoteRequest:
		return s.SignVoteRequest, nil
	case *upstreampb.Message_SignedVoteResponse:
		return s.SignedVoteResponse, nil
	case *upstreampb.Message_SignProposalRequest:
		return s.SignProposalRequest, nil
	case *upstreampb.Message_SignedProposalResponse:
		return s.SignedProposalResponse, nil
	case *upstreampb.Message_PingRequest:
		return s.PingRequest, nil
	case *upstreampb.Message_PingResponse:
		return s.PingResponse, nil
	default:
		return nil, fmt.Errorf("upstream.UnwrapMsg: unknown Sum variant %T", s)
	}
}
