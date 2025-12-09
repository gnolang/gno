package metadata

type Option func(m *Metadata)

// WithWebSocketID sets the WS connection ID
// for the connection metadata
func WithWebSocketID(id string) Option {
	return func(m *Metadata) {
		m.WebSocketID = &id
	}
}
