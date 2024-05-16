package event

import "github.com/gnolang/gno/gno.me/gno"

type Listener struct {
	eventCh chan gno.Event
}
