package conns

import (
	"github.com/olahol/melody"
)

// ConnectionManager defines a connection manager interface
// for active WS connections
type ConnectionManager interface {
	// AddWSConnection registers a new WS connection
	AddWSConnection(id string, session *melody.Session)

	// RemoveWSConnection Removes the WS connection with the supplied ID
	RemoveWSConnection(id string)

	// GetWSConnection fetches a WS connection, if any, using the supplied ID
	GetWSConnection(id string) WSConnection
}

// WSConnection represents a single WS connection
type WSConnection interface {
	// WriteData pushes out data to the WS connection.
	// Returns an error if the write failed (ex. connection closed)
	WriteData(data any) error
}
