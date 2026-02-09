package wsconn

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/conns"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/writer"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/writer/ws"
	"github.com/olahol/melody"
)

// Conns manages active WS connections
type Conns struct {
	logger *slog.Logger
	conns  map[string]Conn // ws connection ID -> conn

	mux sync.RWMutex
}

// NewConns creates a new instance of the WS connection manager
func NewConns(logger *slog.Logger) *Conns {
	return &Conns{
		logger: logger,
		conns:  make(map[string]Conn),
	}
}

// AddWSConnection registers a new WS connection
func (pw *Conns) AddWSConnection(id string, session *melody.Session) {
	pw.mux.Lock()
	defer pw.mux.Unlock()

	ctx, cancelFn := context.WithCancel(context.Background())

	pw.conns[id] = Conn{
		ctx:      ctx,
		cancelFn: cancelFn,
		writer: ws.New(
			pw.logger.With(
				"ws-conn",
				fmt.Sprintf("ws-%s", id),
			),
			session,
		),
	}
}

// RemoveWSConnection removes an existing WS connection
func (pw *Conns) RemoveWSConnection(id string) {
	pw.mux.Lock()
	defer pw.mux.Unlock()

	conn, found := pw.conns[id]
	if !found {
		return
	}

	// Cancel the connection context
	conn.cancelFn()

	delete(pw.conns, id)
}

// GetWSConnection fetches a WS connection, if any
func (pw *Conns) GetWSConnection(id string) conns.WSConnection {
	pw.mux.RLock()
	defer pw.mux.RUnlock()

	conn, found := pw.conns[id]
	if !found {
		return nil
	}

	return &conn
}

// Conn is a single WS connection
type Conn struct {
	ctx      context.Context
	cancelFn context.CancelFunc

	writer writer.ResponseWriter
}

// WriteData writes arbitrary data to the WS connection
func (c *Conn) WriteData(data any) error {
	if c.ctx.Err() != nil {
		return c.ctx.Err()
	}

	c.writer.WriteResponse(data)

	return nil
}
