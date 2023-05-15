package events

import (
	"fmt"

	auto "github.com/gnolang/gno/tm2/pkg/autofile"
)

// StoreStream stores events to disk but is also listenaable.
type StoreStream interface {
	Eventable
	SetHeight(height int64) // to demarcate height in WAL for replay.
}

type storeStream struct {
	afile  *auto.AutoFile
	buf    []byte
	height int64
}

func (ss *storeStream) SetHeight(height int64) {
	if ss.height < height {
		// write new height
		ss.height = height
	} else /* if height <= ss.height */ {
		panic(fmt.Sprintf("invalid SetHeight height value. current %v, got %v", ss.height, height))
	}
}

//----------------------------------------
// move to own file

// FilterStream is listenable and lets you filter.
type FilterStream interface {
	Eventable
}
