package emitter

import "github.com/gnolang/gno/contribs/gnobro/pkg/events"

type NoopServer struct{}

func (*NoopServer) Emit(evt events.Event) {}
