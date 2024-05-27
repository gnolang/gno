package event

import (
	"context"

	"github.com/gnolang/gno/gno.me/state"
)

type Creator interface {
	CreateEvents(
		ctx context.Context,
		appName string,
		functionName string,
		args ...string,
	) (events []*state.Event, err error)
}
