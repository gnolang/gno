package amino

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gnolang/gno/pkgs/amino/pkg"
	"github.com/stretchr/testify/assert"
)

type FooList []*Foo

type Foo struct {
	a string
	b int32
	c []*Foo
	D string // exposed
}

type pair struct {
	Key   string
	Value interface{}
}

func (pr pair) get(key string) (value interface{}) {
	if pr.Key != key {
		panic(fmt.Sprintf("wanted %v but is %v", key, pr.Key))
	}
	return pr.Value
}

func (f Foo) MarshalAmino() ([]pair, error) { //nolint: golint
	return []pair{
		{"a", f.a},
		{"b", f.b},
		{"c", FooList(f.c)},
		{"D", f.D},
	}, nil
}

func (f *Foo) UnmarshalAmino(repr []pair) error {
	f.a = repr[0].get("a").(string)
	f.b = repr[1].get("b").(int32)
	f.c = repr[2].get("c").(FooList)
	f.D = repr[3].get("D").(string)
	return nil
}

var (
	gopkg       = reflect.TypeOf(Foo{}).PkgPath()
	testPackage = pkg.NewPackage(gopkg, "tests", "").
			WithDependencies().
			WithTypes(FooList(nil))
)

func TestMarshalAminoBinary(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterPackage(testPackage)

	f := Foo{
		a: "K",
		b: 2,
		c: []*Foo{{}, {}, {}},
		D: "J",
	}
	bz, err := cdc.MarshalSized(f)
	assert.NoError(t, err)

	t.Logf("bz %#v", bz)

	var f2 Foo
	err = cdc.UnmarshalSized(bz, &f2)
	assert.NoError(t, err)

	assert.Equal(t, f, f2)
	assert.Equal(t, f.a, f2.a) // In case the above doesn't check private fields?
}

func TestMarshalAminoJSON(t *testing.T) {
	cdc := NewCodec()
	cdc.RegisterPackage(testPackage)

	f := Foo{
		a: "K",
		b: 2,
		c: []*Foo{nil, nil, nil},
		D: "J",
	}
	bz, err := cdc.MarshalJSON(f)
	assert.Nil(t, err)

	t.Logf("bz %X", bz)

	var f2 Foo
	err = cdc.UnmarshalJSON(bz, &f2)
	assert.Nil(t, err)

	assert.Equal(t, f, f2)
	assert.Equal(t, f.a, f2.a) // In case the above doesn't check private fields?
}
