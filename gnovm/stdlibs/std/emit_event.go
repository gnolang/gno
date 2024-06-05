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
	}

	_, pkgPath := currentRealm(m)
	fnIdent := getPrevFunctionNameFromTarget(m, "Emit")
	prevFunc := getPrevFunctionNameFromTarget(m, fnIdent) // get the previous function name of the function that called Emit

	prevAddr, prevPath := prevRealm(m)
	prev := prevStack{
		PrevPkgPath: prevPath,
		PrevPkgAddr: prevAddr,
	}

	ctx := GetContext(m)

	evt := gnoEvent{
		OrigCaller: ctx.OrigCaller.String(),
		Prev:       prev,
		PrevFunc:   prevFunc,
		PkgPath:    pkgPath,
		Func:       fnIdent,
		Type:       typ,
		Attributes: eventAttrs,
	}

	ctx.EventLogger.EmitEvent(evt)
}

func attrKeysAndValues(attrs []string) ([]gnoEventAttribute, error) {
	attrLen := len(attrs)
	if attrLen%2 != 0 {
		return nil, errInvalidGnoEventAttrs
	}
	eventAttrs := make([]gnoEventAttribute, attrLen/2)
	for i := 0; i < attrLen-1; i += 2 {
		eventAttrs[i/2] = gnoEventAttribute{
			Key:   attrs[i],
			Value: attrs[i+1],
		}
	}
	return eventAttrs, nil
}

type gnoEvent struct {
	OrigCaller string              `json:"orig_caller"`
	Prev       prevStack           `json:"prev"`
	PrevFunc   string              `json:"prev_func"`
	PkgPath    string              `json:"pkg_path"`
	Func       string              `json:"func"`
	Type       string              `json:"type"`
	Attributes []gnoEventAttribute `json:"attrs"`
}

type prevStack struct {
	PrevPkgPath string `json:"prev_pkg_path"`
	PrevPkgAddr string `json:"prev_pkg_addr"`
}

func (e gnoEvent) AssertABCIEvent() {}

type gnoEventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func prevRealm(m *gno.Machine) (string, string) {
	return X_getRealm(m, 1)
}
