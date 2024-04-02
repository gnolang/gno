package std

// ref: https://github.com/gnolang/gno/pull/853

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

type Event struct {
	Type       string               // Type of event
	PkgPath    string               // Path of the package that emitted the event
	Height     int64                // Height at which the event was emitted
	Timestamp  int64                // Timestamp at which the event was emitted
	Attributes []sdk.EventAttribute // List of event attributes
}

// NewEvent creates a new event with the given type and attributes.
// The attributes must be a list of key-value pairs. The number of attributes
func NewEvent(ctx *ExecContext, typ string, pkgPath string, attrs ...string) (Event, error) {
	if typ == "" {
		return Event{}, fmt.Errorf("event type cannot be empty")
	}

	eventAttrs := make([]sdk.EventAttribute, 0, len(attrs)/2)

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
		PkgPath:    pkgPath,
		Height:     ctx.Height,
		Timestamp:  ctx.Timestamp,
		Attributes: eventAttrs,
	}, nil
}

// AddAttribute adds a new key-value pair to the event attributes.
// It appends the new attribute to the existing list of attributes.
func (e *Event) AddAttribute(key, value string) {
	e.Attributes = append(
		e.Attributes,
		sdk.EventAttribute{Key: key, Value: value},
	)
}

func X_emitEvent(m *gno.Machine, typ string, attrs []string) {
	eventAttrs := make([]sdk.EventAttribute, len(attrs)/2)
	pkgPath := CurrentRealmPath(m)
	event := sdk.NewEvent(typ, pkgPath, eventAttrs...)
	ctx := m.Context.(ExecContext)
	ctx.EventLogger.EmitEvent(event)
}
