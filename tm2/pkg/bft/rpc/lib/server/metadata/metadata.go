package metadata

// Metadata houses the active request metadata
type Metadata struct {
	WebSocketID *string
	RemoteAddr  string
}

// NewMetadata creates a new request metadata object
func NewMetadata(remoteAddr string, opts ...Option) *Metadata {
	m := &Metadata{
		RemoteAddr: remoteAddr,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// IsWS returns a flag indicating if the request
// belongs to a WS connection
func (m *Metadata) IsWS() bool {
	return m.WebSocketID != nil
}
