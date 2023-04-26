package vmk

import (
	"fmt"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	vmh "github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
	"strings"
)

type Dispatcher struct {
	router  sdk.Router
	logger  log.Logger
	icbChan *IBC
}

func NewDispatcher(logger log.Logger) *Dispatcher {
	return &Dispatcher{
		router: sdk.NewRouter(),
		logger: logger.With("vmDispatcher"),
	}
}

// Router returns the router of the BaseApp.
func (d *Dispatcher) Router() sdk.Router {
	return d.router
}

// iterates through all inner messages, route to another infra's contract or another chain's contract
// TODO: not need sdk context here?, another method for handler
func (d *Dispatcher) HandleInternalMsgs(ctx sdk.Context, msgs []vmh.MsgCall, mode sdk.RunTxMode) (result sdk.Result) {
	println("handle internal msgs")
	msgLogs := make([]string, 0, len(msgs))

	data := make([]byte, 0, len(msgs))
	err := error(nil)
	events := []abci.Event{}

	// NOTE: GasWanted is determined by ante handler and GasUsed by the GasMeter.
	for i, msg := range msgs {
		// match message route
		msgRoute := msg.Route()
		handler := d.router.Route(msgRoute)
		if handler == nil {
			result.Error = sdk.ABCIError(std.ErrUnknownRequest("unrecognized message type: " + msgRoute))
			return
		}

		var msgResult sdk.Result
		ctx = ctx.WithEventLogger(sdk.NewEventLogger())

		// run the message!
		// skip actual execution for CheckTx mode
		if mode != sdk.RunTxModeCheck {
			msgResult = handler.Process(ctx, msg)
		}

		println("after process:, msgResult is:", string(msgResult.Data))

		// Each message result's Data must be length prefixed in order to separate
		// each result.
		data = append(data, msgResult.Data...)
		events = append(events, msgResult.Events...)
		// TODO append msgevent from ctx. XXX XXX

		// stop execution and return on first failed message
		if msgResult.Error != nil {
			d.logger.Debug("process innerMsg fail, break and return: ", msgResult.Error.Error())
			msgLogs = append(msgLogs,
				fmt.Sprintf("msg:%d,fail:%v,log:%s,events:%v",
					i, false, msgResult.Log, events))
			err = msgResult.Error
			break
		}

		msgLogs = append(msgLogs,
			fmt.Sprintf("msg:%d,success:%v,log:%s,events:%v",
				i, true, msgResult.Log, events))

		msgLogs = append(msgLogs,
			fmt.Sprintf("msg:%d,success:%v,log:%s,events:%v",
				i, true, msgResult.Log, events))
	}
	// succeed iff all internal msgs succeed and callback succeed
	result.Error = sdk.ABCIError(err)
	result.Data = data
	result.Log = strings.Join(msgLogs, "\n")
	result.GasUsed = ctx.GasMeter().GasConsumed()
	result.Events = events
	return result
}

// send IBC packet, only support gnovm type call (MsgCall) for now, gnoVM <-> gnoVM
func (d *Dispatcher) HandleIBCMsgs(ctx sdk.Context, req vmh.GnoReq) {
	// set callback map

	// this simulates a IBC call, using a chan to loop back
	go d.icbChan.SendPacket(req)
}
