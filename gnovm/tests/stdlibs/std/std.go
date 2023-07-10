package std

import (
	"fmt"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func AssertOriginCall(m *gno.Machine) {
	if !IsOriginCall(m) {
		m.Panic(typedString("invalid non-origin call"))
	}
}

func typedString(s gno.StringValue) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(s)
	return tv
}

func IsOriginCall(m *gno.Machine) bool {
	tname := m.Frames[0].Func.Name
	switch tname {
	case "main": // test is a _filetest
		return len(m.Frames) == 3
	case "runtest": // test is a _test
		return len(m.Frames) == 7
	}
	// support init() in _filetest
	// XXX do we need to distinguish from 'runtest'/_test?
	// XXX pretty hacky even if not.
	if strings.HasPrefix(string(tname), "init.") {
		return len(m.Frames) == 3
	}
	panic("unable to determine if test is a _test or a _filetest")
}

func TestCurrentRealm(m *gno.Machine) string {
	return m.Realm.Path
}

func TestSkipHeights(count int64) {
	panic("not implemented")
	/*
		ctx := m.Context.(stdlibs.ExecContext)
		ctx.Height += count
		m.Context = ctx
	*/
}

func ClearStoreCache(m *gno.Machine) {
	if gno.IsDebug() && testing.Verbose() {
		m.Store.Print()
		fmt.Println("========================================")
		fmt.Println("CLEAR CACHE (RUNTIME)")
		fmt.Println("========================================")
	}
	m.Store.ClearCache()
	m.PreprocessAllFilesAndSaveBlockNodes()
	if gno.IsDebug() && testing.Verbose() {
		m.Store.Print()
		fmt.Println("========================================")
		fmt.Println("CLEAR CACHE DONE")
		fmt.Println("========================================")
	}
}
