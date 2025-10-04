package transfer

import (
	"crypto/sha256"
	"strings"

	"gno.land/p/aib/ibc/host"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/ufmt"
)

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

// IsNative returns true if the denomination is native, thus containing no
// trace history.
func (d Denom) IsNative() bool {
	return len(d.Trace) == 0
}

// Path returns the full denomination according to the ICS20 specification:
// trace + "/" + baseDenom
// If there exists no trace then the base denomination is returned.
func (d Denom) Path() string {
	if d.IsNative() {
		return d.Base
	}

	var sb strings.Builder
	for _, t := range d.Trace {
		sb.WriteString(t.String())
		sb.WriteByte('/')
	}
	sb.WriteString(d.Base)
	return sb.String()
}

// Hash returns the hex bytes of the SHA256 hash of the Denom fields using the
// following formula:
//
// hash = sha256(trace + "/" + baseDenom)
func (d Denom) Hash() []byte {
	hash := sha256.Sum256([]byte(d.Path()))
	return hash[:]
}

// IBCDenom a coin denomination for an ICS20 fungible token in the format
// 'ibc/{hash(trace + baseDenom)}'. If the trace is empty, it will return the
// base denomination.
func (d Denom) IBCDenom() string {
	if d.IsNative() {
		return d.Base
	}

	return ufmt.Sprintf("%s/%s", DenomPrefix, d.Hash())
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
