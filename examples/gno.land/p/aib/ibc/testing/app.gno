package testing

import (
	"chain/banker"

	"gno.land/p/aib/ibc/app"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/ufmt"
)

type Args struct {
	SourceClient      string
	DestinationClient string
	Sequence          uint64
	Payload           types.Payload
}

type App struct {
	Calls struct {
		OnSendPacket                  []Args
		OnSendPacketReturn            error
		OnRecvPacket                  []Args
		OnRecvPacketReturn            types.RecvPacketResult
		OnTimeoutPacket               []Args
		OnTimeoutPacketReturn         error
		OnAcknowledgementPacket       []Args
		OnAcknowledgementPacketReturn error
	}
}

var _ app.IBCApp = &App{}

// Implements app.IBCApp
func (a *App) OnSendPacket(
	_ banker.Banker,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) error {
	a.Calls.OnSendPacket = append(a.Calls.OnSendPacket,
		Args{
			SourceClient:      sourceClient,
			DestinationClient: destinationClient,
			Sequence:          sequence,
			Payload:           payload,
		})
	return a.Calls.OnSendPacketReturn
}

// Implements app.IBCApp
func (a *App) OnRecvPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) types.RecvPacketResult {
	a.Calls.OnRecvPacket = append(a.Calls.OnRecvPacket,
		Args{
			SourceClient:      sourceClient,
			DestinationClient: destinationClient,
			Sequence:          sequence,
			Payload:           payload,
		})
	return a.Calls.OnRecvPacketReturn
}

// Implements app.IBCApp
func (a *App) OnTimeoutPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) error {
	a.Calls.OnTimeoutPacket = append(a.Calls.OnTimeoutPacket,
		Args{
			SourceClient:      sourceClient,
			DestinationClient: destinationClient,
			Sequence:          sequence,
			Payload:           payload,
		})
	return a.Calls.OnTimeoutPacketReturn
}

// Implements app.IBCApp
func (a *App) OnAcknowledgementPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	acknowledgement []byte,
	payload types.Payload,
) error {
	a.Calls.OnAcknowledgementPacket = append(a.Calls.OnAcknowledgementPacket,
		Args{
			SourceClient:      sourceClient,
			DestinationClient: destinationClient,
			Sequence:          sequence,
			Payload:           payload,
		})
	return a.Calls.OnAcknowledgementPacketReturn
}

func (a *App) Report() string {
	var s string
	s += ufmt.Sprintf("OnSendPacket (%d)\n", len(a.Calls.OnSendPacket))
	for _, args := range a.Calls.OnSendPacket {
		s += args.Report() + "\n"
	}
	s += ufmt.Sprintf("OnRecvPacket (%d)\n", len(a.Calls.OnRecvPacket))
	for _, args := range a.Calls.OnRecvPacket {
		s += args.Report() + "\n"
	}
	s += ufmt.Sprintf("OnTimeoutPacket (%d)\n", len(a.Calls.OnTimeoutPacket))
	for _, args := range a.Calls.OnTimeoutPacket {
		s += args.Report() + "\n"
	}
	s += ufmt.Sprintf("OnAcknowledgementPacket (%d)\n", len(a.Calls.OnAcknowledgementPacket))
	for _, args := range a.Calls.OnAcknowledgementPacket {
		s += args.Report() + "\n"
	}
	return s
}

func (a *Args) Report() string {
	var s string
	s += ufmt.Sprintf("- sourceClient: %s\n", a.SourceClient)
	s += ufmt.Sprintf("- destinationClient: %s\n", a.DestinationClient)
	s += ufmt.Sprintf("- sequence: %d\n", a.Sequence)
	s += "- payload:\n"
	s += ufmt.Sprintf("  - sourcePort: %s\n", a.Payload.SourcePort)
	s += ufmt.Sprintf("  - destinationPort: %s\n", a.Payload.DestinationPort)
	s += ufmt.Sprintf("  - version: %s\n", a.Payload.Version)
	s += ufmt.Sprintf("  - encoding: %s\n", a.Payload.Encoding)
	s += ufmt.Sprintf("  - value: %s\n", string(a.Payload.Value))
	return s
}
