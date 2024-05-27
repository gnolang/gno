package server

import (
	"context"

	"github.com/gnolang/gno/gno.me/gno"
	"github.com/gnolang/gno/gno.me/state"
)

type eventCreator struct {
	vm gno.VM
}

func (c eventCreator) CreateEvents(
	ctx context.Context,
	appName string,
	functionName string,
	args ...string,
) ([]*state.Event, error) {
	_, events, err := c.vm.Call(ctx, appName, false, functionName, args...)
	return events, err
}
