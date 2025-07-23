// PKGPATH: gno.land/r/evt
package evt

type Event struct {
	name string
}

var deletionEvents = []*Event{
	{name: "event1"},
	{name: "event2"},
	{name: "event3"},
	{name: "event4"},
}

var insertEvents = []*Event{
	{name: "event1"},
	{name: "event2"},
}

var appendEvents = []*Event{
	{name: "event1"},
}

func DelEvent(name string) {
	for i, event := range deletionEvents {
		if event.name == name {
			deletionEvents = append(deletionEvents[:i], deletionEvents[i+1:]...)
			return
		}
	}
}

func InsertEvent(name string) {
	insertEvents = append(insertEvents[:1], append([]*Event{{name: name}}, insertEvents[1:]...)...)
}

func AppendEvent(name string) {
	appendEvents = append(appendEvents, &Event{name: name})
}

func printEvents(events []*Event) {
	for _, event := range events {
		println(event.name)
	}
}

func main(cur realm) {
	DelEvent("event2")
	InsertEvent("event1.5")
	AppendEvent("event2")

	printEvents(deletionEvents)
	printEvents(insertEvents)
	printEvents(appendEvents)
}

// Output:
// event1
// event3
// event4
// event1
// event1.5
// event2
// event1
// event2
