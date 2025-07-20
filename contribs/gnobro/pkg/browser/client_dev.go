package browser

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gnolang/gno/contribs/gnodev/pkg/emitter"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gorilla/websocket"
)

type ClientStatus int

const (
	ClientStatusDisconnected = iota
	ClientStatusConnecting
	ClientStatusConnected
)

const MaxBackoff = time.Second * 20

var ErrHandlerNotSet = errors.New("handler not set")

type DevClient struct {
	Logger        *slog.Logger
	Handler       func(typ events.Type, data any) error
	HandlerStatus func(status ClientStatus, remote string)

	conn *websocket.Conn
}

func (c *DevClient) updateStatus(s ClientStatus, remote string) {
	if c.HandlerStatus != nil {
		c.HandlerStatus(s, remote)
	}
}

func (c *DevClient) Run(ctx context.Context, addr string, header http.Header) error {
	if c.Handler == nil {
		return ErrHandlerNotSet
	}

	defer c.updateStatus(ClientStatusDisconnected, addr)
	c.updateStatus(ClientStatusDisconnected, addr)

	if c.Logger == nil {
		c.Logger = log.NewNoopLogger()
	}

	for ctx.Err() == nil {
		c.updateStatus(ClientStatusConnecting, addr)
		if err := c.dialBackoff(ctx, addr, nil); err != nil {
			return err
		}

		c.updateStatus(ClientStatusConnected, addr)

		c.Logger.Info("connected to server", "addr", addr)

		err := c.handleEvents(ctx)
		if err == nil {
			return nil
		}

		c.updateStatus(ClientStatusDisconnected, addr)

		var closeError *websocket.CloseError
		if errors.As(err, &closeError) {
			c.Logger.Error("connection has been closed, reconnecting...", "err", closeError)
			continue
		}

		return fmt.Errorf("unexpected error: %w", err)
	}

	return context.Cause(ctx)
}

func (c *DevClient) dialBackoff(ctx context.Context, addr string, header http.Header) error {
	dialer := websocket.DefaultDialer
	backoff := time.Second
	for {
		var err error

		c.Logger.Debug("connecting to dev events endpoint", "addr", addr)
		c.conn, _, err = dialer.DialContext(ctx, addr, header)
		if ctx.Err() != nil {
			return context.Cause(ctx)
		}

		if err == nil {
			return nil
		}

		switch {
		case backoff > MaxBackoff:
			backoff = MaxBackoff
		case backoff < MaxBackoff:
			backoff *= 2
		default:
		}

		c.Logger.Info("could not connect to server", "err", err, "next_attempt", backoff)
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(backoff):
		}
	}
}

func (c *DevClient) handleEvents(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		c.conn.Close()
	}()

	for {
		var evt emitter.EventJSON
		if err := c.conn.ReadJSON(&evt); err != nil {
			if ctx.Err() != nil {
				return nil
			}

			return fmt.Errorf("unable to read json event: %w", err)
		}

		c.Logger.Info("receiving event", "evt", evt.Type)

		if err := c.Handler(evt.Type, evt.Data); err != nil {
			return fmt.Errorf("unable to handle event: %w", err)
		}
	}
}
