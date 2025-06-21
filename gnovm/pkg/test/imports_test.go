// imports_test.go
// Blackbox tests for LoadImports(): inject panics of various types via store.GetPackage
// to ensure LoadImports recovers them into appropriate errors.

package test_test

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"

	// "runtime/debug"
	"strings"
	"testing"
	"unsafe"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	testpkg "github.com/gnolang/gno/gnovm/pkg/test"

	// teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	// "github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	// "github.com/gnolang/gno/tm2/pkg/db/memdb"
	tm2std "github.com/gnolang/gno/tm2/pkg/std"
)

// makeMemPkg returns a MemPackage that imports "foo" so LoadImports calls GetPackage.
func makeMemPkg() *tm2std.MemPackage {
	return &tm2std.MemPackage{
		Name:  "pkg",
		Path:  "pkg",
		Files: []*tm2std.MemFile{{Name: "a.gno", Body: "package pkg\nimport \"foo\"\n"}},
	}
}

// makeStoreThatPanics builds a test store and overrides GetPackage to panic val.
func makeStoreThatPanics(val interface{}) gno.Store {
	_, realStore := testpkg.Store("", io.Discard)
	realStore.SetPackageGetter(func(pkgPath string, create gno.Store) (*gno.PackageNode, *gno.PackageValue) {
		panic(val)
	})
	return realStore
}

// newPreprocessError constructs a *gno.PreprocessError wrapping inner, via reflection.
func newPreprocessError(inner error) *gno.PreprocessError {
	tp := reflect.TypeOf((*gno.PreprocessError)(nil)).Elem()
	v := reflect.New(tp)
	pe := v.Interface().(*gno.PreprocessError)
	// set unexported 'err' field
	f := v.Elem().FieldByName("err")
	f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	f.Set(reflect.ValueOf(inner))
	return pe
}

func TestLoadImports_RecoveryTypedValue(t *testing.T) {
	val := &gno.TypedValue{T: nil, V: gno.BigintValue{V: big.NewInt(7)}}
	store := makeStoreThatPanics(val)
	err := testpkg.LoadImports(store, makeMemPkg(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != val.String() {
		t.Errorf("got %q; want %q", err.Error(), val.String())
	}
}

func TestLoadImports_RecoveryPreprocessError(t *testing.T) {
	inner := errors.New("bad preprocess")
	val := newPreprocessError(inner)
	store := makeStoreThatPanics(val)
	err := testpkg.LoadImports(store, makeMemPkg(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, inner) {
		t.Errorf("expected error to wrap %v; got %v", inner, err)
	}
}

func TestLoadImports_RecoveryUnhandledPanicError(t *testing.T) {
	expectedErrDescriptor := "oops!"
	errVal := gno.UnhandledPanicError{Descriptor: expectedErrDescriptor}
	store := makeStoreThatPanics(errVal)
	err := testpkg.LoadImports(store, makeMemPkg(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errVal) {
		t.Errorf("expected error to be %#v; got %#v", errVal, err)
	}
}

func TestLoadImports_RecoveryGenericError(t *testing.T) {
	expectedErrString := "generic error"
	store := makeStoreThatPanics(errors.New(expectedErrString))
	err := testpkg.LoadImports(store, makeMemPkg(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), expectedErrString) {
		t.Errorf("expected error to contain string %#v; got %#v", expectedErrString, err)
	}
}

func TestLoadImports_RecoveryDefaultPanic(t *testing.T) {
	val := 12345
	store := makeStoreThatPanics(val)
	err := testpkg.LoadImports(store, makeMemPkg(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("%v", val)) {
		t.Errorf("error message %q does not contain %v", err.Error(), val)
	}
	if !strings.Contains(err.Error(), "12345") {
		t.Errorf("expected stack trace in error message: %#v (%#v)", err, err.Error())
	}
}
