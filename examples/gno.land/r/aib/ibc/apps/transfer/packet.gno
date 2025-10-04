package transfer

import (
	"strings"

	"gno.land/p/aib/encoding/proto"
	"gno.land/p/nt/ufmt"
	"gno.land/p/onbloc/int256"
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

// NewFungibleTokenPacketData constructs a new FungibleTokenPacketData instance
func NewFungibleTokenPacketData(
	denom string, amount string,
	sender, receiver string,
	memo string,
) FungibleTokenPacketData {
	return FungibleTokenPacketData{
		Denom:    denom,
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Memo:     memo,
	}
}

// ProtoUnmarshal reads the protobuf encoded bz into f.
//
//	message FungibleTokenPacketData {
//		string denom = 1;
//		string amount = 2;
//		string sender = 3;
//		string receiver = 4;
//		string memo = 5;
//	}
func (f *FungibleTokenPacketData) ProtoUnmarshal(bz []byte) error {
	var pos int
	for pos < len(bz) {
		// Read field tag (key)
		key, newPos, err := proto.DecodeVarint(bz, pos)
		if err != nil {
			return ufmt.Errorf("error reading field tag: %v", err)
		}
		pos = newPos

		fieldNum := key >> 3
		wireType := key & 0x07

		if wireType != 2 { // All fields are strings (wire type 2)
			return ufmt.Errorf("unexpected wire type %d for field %d", wireType, fieldNum)
		}

		// Read string value
		value, newPos, err := proto.DecodeString(bz, pos)
		if err != nil {
			return ufmt.Errorf("error reading string value: %v", err)
		}
		pos = newPos

		// Assign to appropriate field
		switch fieldNum {
		case 1:
			f.Denom = value
		case 2:
			f.Amount = value
		case 3:
			f.Sender = value
		case 4:
			f.Receiver = value
		case 5:
			f.Memo = value
		default:
			return ufmt.Errorf("unknown field aren't authorized in FungibleTokenPacketData")
		}
	}
	return nil
}

// ValidateBasic is used for validating the token transfer.
func (f FungibleTokenPacketData) ValidateBasic() error {
	amount, err := int256.FromDecimal(f.Amount)
	if err != nil {
		return ufmt.Errorf("unable to parse transfer amount %q: %v", f.Amount, err)
	}
	if amount.Sign() != 1 {
		return ufmt.Errorf("amount must be strictly positive: got %s", amount)
	}
	if strings.TrimSpace(f.Sender) == "" {
		return ufmt.Errorf("sender address cannot be blank")
	}
	if strings.TrimSpace(f.Receiver) == "" {
		return ufmt.Errorf("receiver address cannot be blank")
	}
	denom := ExtractDenomFromPath(f.Denom)
	return denom.ValidateBasic()
}
