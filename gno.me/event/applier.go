package event

import (
	"context"

	"github.com/gnolang/gno/gno.me/state"
)

type Applier interface {
	ApplyEvent(ctx context.Context, event *state.Event) error
}
