package events

import "fmt"

type customEvent struct {
	name string
}

// Custom create a new event with the given name, prefixing it with "CUSTOM_".
// If the name is empty, it will panic.
func Custom(name string) Event {
	if name == "" {
		panic("custom event cannot have an empty name")
	}

	return &customEvent{name: fmt.Sprintf("CUSTOM_%s", name)}
}

func (customEvent) assertEvent() {}

func (c *customEvent) Type() Type {
	return Type(c.name)
}
