package transfer

import (
	"gno.land/p/aib/encoding/proto"
	"gno.land/p/nt/ufmt"
)

// FungibleTokenPacketData defines a struct for the packet payload
// See FungibleTokenPacketData spec:
// https://github.com/cosmos/ibc/tree/master/spec/app/ics-020-fungible-token-transfer#data-structures
type FungibleTokenPacketData struct {
	// the token denomination to be transferred
	Denom string
	// the token amount to be transferred
	Amount string
	// the sender address
	Sender string
	// the recipient address on the destination chain
	Receiver string
	// optional memo
	Memo string
}

// DecodeFungibleTokenPacketData decodes the protobuf message from bytes
//
//	message FungibleTokenPacketData {
//		string denom = 1;
//		string amount = 2;
//		string sender = 3;
//		string receiver = 4;
//		string memo = 5;
//	}
func DecodeFungibleTokenPacketData(bz []byte) (*FungibleTokenPacketData, error) {
	var (
		data = &FungibleTokenPacketData{}
		pos  int
	)
	for pos < len(bz) {
		// Read field tag (key)
		key, newPos, err := proto.DecodeVarint(bz, pos)
		if err != nil {
			return nil, ufmt.Errorf("error reading field tag: %v", err)
		}
		pos = newPos

		fieldNum := key >> 3
		wireType := key & 0x07

		if wireType != 2 { // All fields are strings (wire type 2)
			return nil, ufmt.Errorf("unexpected wire type %d for field %d", wireType, fieldNum)
		}

		// Read string value
		value, newPos, err := proto.DecodeString(bz, pos)
		if err != nil {
			return nil, ufmt.Errorf("error reading string value: %v", err)
		}
		pos = newPos

		// Assign to appropriate field
		switch fieldNum {
		case 1:
			data.Denom = value
		case 2:
			data.Amount = value
		case 3:
			data.Sender = value
		case 4:
			data.Receiver = value
		case 5:
			data.Memo = value
		default:
			// Skip unknown fields (proto3 behavior)
			continue
		}
	}
	return data, nil
}
