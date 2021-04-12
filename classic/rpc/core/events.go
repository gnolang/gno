package core

import (
	"fmt"

	"github.com/tendermint/classic/libs/events"
	ctypes "github.com/tendermint/classic/rpc/core/types"
	rpctypes "github.com/tendermint/classic/rpc/lib/types"
)

// NOTE: These websockets are synchronous and blocking.  They will be replaced
// with another implementation of event subscription.
func Subscribe(ctx *rpctypes.Context) (*ctypes.ResultSubscribe, error) {
	addr := ctx.RemoteAddr()
	listenerID := fmt.Sprintf("rpc-subscribe-%v", addr)
	logger.Info("Subscribe to events", "remote", addr)

	eventCh := events.Subscribe(evsw, listenerID)

	go func() {
		for {
			select {
			case event, ok := <-eventCh:
				if ok {
					resultEvent := &ctypes.ResultEvent{Event: event}
					ctx.WSConn.TryWriteRPCResponse(
						rpctypes.NewRPCSuccessResponse(
							ctx.WSConn.Codec(),
							rpctypes.JSONRPCStringID(fmt.Sprintf("%v#event", ctx.JSONReq.ID)),
							resultEvent,
						))
				} else {
					ctx.WSConn.TryWriteRPCResponse(
						rpctypes.RPCServerError(rpctypes.JSONRPCStringID(
							fmt.Sprintf("%v#event", ctx.JSONReq.ID)),
							fmt.Errorf("subscription was aborted"),
						))
				}
			}
		}
	}()

	return &ctypes.ResultSubscribe{}, nil
}

func UnsubscribeAll(ctx *rpctypes.Context) (*ctypes.ResultUnsubscribe, error) {
	addr := ctx.RemoteAddr()
	listenerID := fmt.Sprintf("rpc-subscribe-%v", addr)
	logger.Info("Unsubscribe from all", "remote", addr)

	evsw.RemoveListener(listenerID)
	return &ctypes.ResultUnsubscribe{}, nil
}
