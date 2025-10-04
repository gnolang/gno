package transfer

import (
	"strings"

	"gno.land/p/aib/encoding/proto"
	"gno.land/p/aib/ibc/host"
	"gno.land/p/aib/ibc/types"
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

// Denom holds the base denom of a Token and a trace of the chains it was sent
// through.
type Denom struct {
	// the base token denomination
	Base string
	// the trace of the token
	Trace []Hop
}

// NewDenom creates a new Denom instance given the base denomination and a
// variable number of hops.
func NewDenom(base string, trace ...Hop) Denom {
	return Denom{
		Base:  base,
		Trace: trace,
	}
}

// ValidateBasic performs a basic validation of the Denom fields.
func (d Denom) ValidateBasic() error {
	// NOTE: base denom validation cannot be performed as each chain may define
	// its own base denom validation
	if strings.TrimSpace(d.Base) == "" {
		return ufmt.Errorf("base denomination cannot be blank")
	}

	for _, hop := range d.Trace {
		if err := hop.ValidateBasic(); err != nil {
			return ufmt.Errorf("invalid trace")
		}
	}

	return nil
}

// ExtractDenomFromPath returns the denom from the full path.
func ExtractDenomFromPath(fullPath string) Denom {
	denomSplit := strings.Split(fullPath, "/")

	if denomSplit[0] == fullPath {
		return Denom{
			Base: fullPath,
		}
	}

	var (
		trace          []Hop
		baseDenomSlice []string
	)

	length := len(denomSplit)
	for i := 0; i < length; i += 2 {
		// The IBC specification does not guarantee the expected format of the
		// destination port or destination channel identifier. A short term
		// solution to determine base denomination is to expect the channel
		// identifier to be the one ibc-go specifies. A longer term solution is to
		// separate the path and base denomination in the ICS20 packet. If an
		// intermediate hop prefixes the full denom with a channel identifier
		// format different from our own, the base denomination will be incorrectly
		// parsed, but the token will continue to be treated correctly as an IBC
		// denomination. The hash used to store the token internally on our chain
		// will be the same value as the base denomination being correctly parsed.
		// nolint:revive // Early return possible, but not needed here
		if i < length-1 && length > 2 && types.IsValidClientID(denomSplit[i+1]) {
			trace = append(trace, NewHop(denomSplit[i], denomSplit[i+1]))
		} else {
			baseDenomSlice = denomSplit[i:]
			break
		}
	}

	base := strings.Join(baseDenomSlice, "/")

	return Denom{
		Base:  base,
		Trace: trace,
	}
}

// Hop defines a port ID, channel ID pair specifying a unique "hop" in a trace
type Hop struct {
	PortId string
	// TODO replace channelID by clientiD ?
	ChannelId string
}

// NewHop creates a Hop with the given port ID and channel ID.
func NewHop(portID, channelID string) Hop {
	return Hop{portID, channelID}
}

// ValidateBasic performs a basic validation of the Hop fields.
func (h Hop) ValidateBasic() error {
	if err := host.PortIdentifierValidator(h.PortId); err != nil {
		return ufmt.Errorf("invalid hop source port ID %s", h.PortId)
	}
	if err := host.ChannelIdentifierValidator(h.ChannelId); err != nil {
		return ufmt.Errorf("invalid hop source channel ID %s", h.ChannelId)
	}

	return nil
}

// String returns the Hop in the format:
// <portID>/<channelID>
func (h Hop) String() string {
	return ufmt.Sprintf("%s/%s", h.PortId, h.ChannelId)
}
