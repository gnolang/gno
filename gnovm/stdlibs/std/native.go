package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func IsOriginCall(m *gno.Machine) bool {
	return len(m.Frames) == 2
}

func CurrentRealmPath(m *gno.Machine) string {
	if m.Realm != nil {
		return m.Realm.Path
	}
	return ""
}

func GetChainID(m *gno.Machine) string {
	return m.Context.(execContext).GetChainID()
}

func GetHeight(m *gno.Machine) int64 {
	return m.Context.(execContext).GetHeight()
}

func typedString(s gno.StringValue) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(s)
	return tv
}

type execContext interface {
	GetHeight() int64
	GetChainID() string
}
