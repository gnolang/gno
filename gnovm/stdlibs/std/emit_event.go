package std

// ref: https://github.com/gnolang/gno/pull/575
// ref: https://github.com/gnolang/gno/pull/1833

import (
	"errors"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

var errInvalidGnoEventAttrs = errors.New("cannot pair attributes due to odd count")

func X_emit(m *gno.Machine, typ string, attrs []string) {
	eventAttrs, err := attrKeysAndValues(attrs)
	if err != nil {
		m.Panic(typedString(err.Error()))
		return
	}

	pkgPath := currentPkgPath(m)

	ctx := GetContext(m)

	evt := GnoEvent{
		Type:       typ,
		Attributes: eventAttrs,
		PkgPath:    pkgPath,
	}

	ctx.EventLogger.EmitEvent(evt)
}

func attrKeysAndValues(attrs []string) ([]GnoEventAttribute, error) {
	attrLen := len(attrs)
	if attrLen%2 != 0 {
		return nil, errInvalidGnoEventAttrs
	}
	eventAttrs := make([]GnoEventAttribute, attrLen/2)
	for i := 0; i < attrLen-1; i += 2 {
		eventAttrs[i/2] = GnoEventAttribute{
			Key:   attrs[i],
			Value: attrs[i+1],
		}
	}
	return eventAttrs, nil
}

// XXX rename to std/events.Event?
type GnoEvent struct {
	Type       string              `json:"type"`
	Attributes []GnoEventAttribute `json:"attrs"`
	PkgPath    string              `json:"pkg_path"`
}

func (e GnoEvent) AssertABCIEvent() {}

// XXX rename to std/events.Attribute?
type GnoEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FindAttribute returns the value for the key in e.Attributes. If not found, return "", false
func (e *GnoEvent) FindAttribute(key string) (string, bool) {
	for _, attr := range e.Attributes {
		if attr.Key == key {
			return attr.Value, true
		}
	}

	return "", false
}
