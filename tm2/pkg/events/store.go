package events

// StoreStream stores events to disk but is also listenaable.
type StoreStream interface {
	Eventable
	SetHeight(height int64) // to demarcate height in WAL for replay.
}

// ----------------------------------------
// move to own file

// FilterStream is listenable and lets you filter.
type FilterStream interface {
	Eventable
}
