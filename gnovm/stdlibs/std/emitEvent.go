package std

// ref: https://github.com/gnolang/gno/pull/853

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/sdk"
)

type Event struct {
	Type       string
	Attributes []sdk.EventAttribute
}

func NewEvent(typ string, attrs ...string) (Event, error) {
	var eventAttrs []sdk.EventAttribute

	attrLen := len(attrs)
	if attrLen%2 != 0 {
		return Event{}, fmt.Errorf("attributes has an odd number of elements. current length: %d", attrLen)
	}

	for i := 0; i < attrLen-1; i += 2 {
		eventAttrs = append(eventAttrs, sdk.EventAttribute{
			Key:   attrs[i],
			Value: attrs[i+1],
		})
	}

	return Event{
		Type:       typ,
		Attributes: eventAttrs,
	}, nil
}

func (e *Event) AddAttribute(key, value string) *Event {
	e.Attributes = append(
		e.Attributes,
		sdk.EventAttribute{Key: key, Value: value},
	)

	return e
}
