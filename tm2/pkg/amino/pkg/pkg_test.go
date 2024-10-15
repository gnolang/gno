package pkg

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Foo struct {
	FieldA int
	FieldB string
}

func TestNewPackage(t *testing.T) {
	t.Parallel()

	// This should panic, as slashes in p3pkg is not allowed.
	assert.Panics(t, func() {
		NewPackage("foobar.com/some/path", "some/path", "").WithTypes(Foo{})
	}, "slash in p3pkg should not be allowed")

	// This should panic, as the go pkg path includes a dot in the wrong place.
	assert.Panics(t, func() {
		NewPackage("blah/foobar.com/some/path", "some.path", "").WithTypes(Foo{})
	}, "invalid go pkg path")

	// This should panic, as the go pkg path includes a leading slash.
	assert.Panics(t, func() {
		NewPackage("/foobar.com/some/path", "some.path", "").WithTypes(Foo{})
	}, "invalid go pkg path")

	// This should panic, as the dirname is relative.
	assert.Panics(t, func() {
		NewPackage("foobar.com/some/path", "some.path", "../someplace").WithTypes(Foo{})
	}, "invalid dirname")

	pkg := NewPackage("foobar.com/some/path", "some.path", "")
	assert.NotNil(t, pkg)
}

func TestFullNameForType(t *testing.T) {
	t.Parallel()

	// The Go package depends on how this test is invoked.
	// Sometimes it is "github.com/gnolang/gno/tm2/pkg/amino/packagepkg_test".
	// Sometimes it is "command-line-arguments"
	// Sometimes it is "command-line-arguments_test"
	gopkg := reflect.TypeOf(Foo{}).PkgPath()
	pkg := NewPackage(gopkg, "some.path", "").WithTypes(Foo{})

	assert.Equal(t, "some.path.Foo", pkg.FullNameForType(reflect.TypeOf(Foo{})))

	typeURL := pkg.TypeURLForType(reflect.TypeOf(Foo{}))
	assert.False(t, strings.Contains(typeURL[1:], "/"))
	assert.Equal(t, "/", string(typeURL[0]))
}

// If the struct wasn't registered, you can't get a name or type_url for it.
func TestFullNameForUnexpectedType(t *testing.T) {
	t.Parallel()

	gopkg := reflect.TypeOf(Foo{}).PkgPath()
	pkg := NewPackage(gopkg, "some.path", "")

	assert.Panics(t, func() {
		pkg.FullNameForType(reflect.TypeOf(Foo{}))
	})

	assert.Panics(t, func() {
		pkg.TypeURLForType(reflect.TypeOf(Foo{}))
	})
}
