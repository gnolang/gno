package emitter

import "github.com/gnolang/gno/contribs/gnodev/pkg/events"

type NoopServer struct{}

func (*NoopServer) Emit(evt events.Event) {}
