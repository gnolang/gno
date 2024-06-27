// PKGPATH: gno.land/r/evt_test
package evt_test

type Event struct {
	name string
}

var events = []*Event{
	{name: "event1"},
	{name: "event2"},
	{name: "event3"},
	{name: "event4"},
}

func DelEvent(name string) {
	for i, event := range events {
		if event.name == name {
			events = append(events[:i], events[i+1:]...)
			return
		}
	}
}

func main() {
	DelEvent("event2")
}

// No output for this test. It passes if it does not panic.
