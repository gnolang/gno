package vm

import (
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// It returns a AnteHandler function that can be chained togateher in app to check pre condition before put it the mempool and propergate to other nodes.
// It checks  there is enough gas to execute the transaction in TxCheck and Simulation mode.
// XXX: We only abort the tx due to the insufficient gas. Should we even allow it pass ante handler to prevent censorship? In other word, should we only keep the min gas price as the check to drop a transaction?
func NewAnteHandler(vmKpr *VMKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx std.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {
		// skip the check for Deliver Mode Gas and Msg Executions
		if ctx.Mode() == sdk.RunTxModeDeliver {
			return ctx, res, false
		}
		// XXXX: check vm gas here for CheckTx and Simulation node.

		vmh := NewHandler(vmKpr)
		msgs := tx.GetMsgs()

		for _, msg := range msgs {
			// match message route
			msgRoute := msg.Route()
			if msgRoute == RouterKey {
				// XXX: When there is no enough gas left in gas meter, it will stop the transaction in CheckTx() before passing the tx to mempool and broadcasting to other nodes.
				// Same message will be processed sencond time in DeliverTx(). It should be ok for now, since CheckTx and DeliverTx execution are in different contexts that are not linked to each other.

				res = vmh.Process(ctx, msg)
			}

			// we don't abort the transaction when there is a message execution failure. The failed message should be allowed to propagate to other nodes.
			// We dont not want to censor the tx for VM execution failures just by one node.
			// XXX: Do not uncomment this. Do not remvove this either to prevent someone accidentally add this check.
			// if !res.IsOK() {
			//	return ctx, res, true
			// }
		}

		return ctx, res, false
	}
}
