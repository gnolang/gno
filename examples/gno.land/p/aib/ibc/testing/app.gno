package testing

import (
	"gno.land/p/aib/ibc/app"
	"gno.land/p/aib/ibc/types"
)

type App struct {
	Calls struct {
		OnSendPacket            int
		OnRecvPacket            int
		OnTimeoutPacket         int
		OnAcknowledgementPacket int
	}
}

var _ app.IBCApp = &App{}

// Implements app.IBCApp
func (a *App) OnSendPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) error {
	a.Calls.OnSendPacket++
	return nil
}

// Implements app.IBCApp
func (a *App) OnRecvPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) types.RecvPacketResult {
	a.Calls.OnRecvPacket++
	return types.RecvPacketResult{}
}

// Implements app.IBCApp
func (a *App) OnTimeoutPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload types.Payload,
) error {
	a.Calls.OnTimeoutPacket++
	return nil
}

// Implements app.IBCApp
func (a *App) OnAcknowledgementPacket(
	sourceClient string,
	destinationClient string,
	sequence uint64,
	acknowledgement []byte,
	payload types.Payload,
) error {
	a.Calls.OnAcknowledgementPacket++
	return nil
}
