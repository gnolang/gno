package events

type EventType string

const (
	EvtReload         EventType = "NODE_RELOAD"
	EvtReset          EventType = "NODE_RESET"
	EvtPackagesUpdate EventType = "PACKAGES_UPDATE"
)

type Event struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

// Event Reload

type EventReload struct{}

func NewEventReload() *Event {
	return &Event{
		Type: EvtReload,
		Data: &EventReload{},
	}
}

// Event Reset

type EventReset struct{}

func NewEventReset() *Event {
	return &Event{
		Type: EvtReload,
		Data: &EventReset{},
	}
}

// Event Packages Update

type PackageUpdate struct {
	Package string   `json:"package"`
	Files   []string `json:"files"`
}

type PackagesUpdateEvent struct {
	Pkgs []PackageUpdate `json:"packages"`
}

func NewPackagesUpdateEvent(pkgs []PackageUpdate) *Event {
	return &Event{
		Type: EvtPackagesUpdate,
		Data: &PackagesUpdateEvent{
			Pkgs: pkgs,
		},
	}
}
