package components

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateFunc_AddSub(t *testing.T) {
	t.Parallel()

	add := funcMap["add"].(func(int, int) int)
	sub := funcMap["sub"].(func(int, int) int)

	assert.Equal(t, 5, add(2, 3))
	assert.Equal(t, 0, add(0, 0))
	assert.Equal(t, -1, add(1, -2))

	assert.Equal(t, 1, sub(3, 2))
	assert.Equal(t, -1, sub(0, 1))
}

func TestTemplateFunc_DerefInt(t *testing.T) {
	t.Parallel()

	derefInt := funcMap["derefInt"].(func(*int) int)

	v := 42
	assert.Equal(t, 42, derefInt(&v))
	assert.Equal(t, 0, derefInt(nil), "nil pointer should return 0")
}

func TestTemplateFunc_HeadingForKind(t *testing.T) {
	t.Parallel()

	headingForKind := funcMap["headingForKind"].(func(string) string)

	assert.Equal(t, "Key", headingForKind(KindMap))
	assert.Equal(t, "Index", headingForKind(KindSlice))
	assert.Equal(t, "Index", headingForKind(KindArray))
	assert.Equal(t, "Field", headingForKind(KindStruct))
	assert.Equal(t, "Field", headingForKind("unknown"))
}

func TestTemplateFunc_KindIconID(t *testing.T) {
	t.Parallel()

	kindIconID := funcMap["kindIconID"].(func(string, string) string)

	cases := []struct {
		kind, typ string
		want      string
	}{
		{KindPrimitive, "string", "kind-string"},
		{KindPrimitive, "bool", "kind-bool"},
		{KindPrimitive, "int64", "kind-number"},
		{KindStruct, "", "kind-struct"},
		{KindMap, "", "kind-map"},
		{KindSlice, "", "kind-slice"},
		{KindArray, "", "kind-slice"},
		{KindPointer, "", "kind-pointer"},
		{KindFunc, "", "kind-func"},
		{KindClosure, "", "kind-closure"},
		{KindRef, "", "kind-ref"},
		{KindNil, "", "kind-nil"},
		{KindPackage, "", "kind-package"},
		{KindType, "", "kind-type"},
		{KindInterface, "", "kind-interface"},
		{"unknown", "", "kind-unknown"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, kindIconID(tc.kind, tc.typ),
			"kind=%q typ=%q", tc.kind, tc.typ)
	}
}

func TestTemplateFunc_KindGroup(t *testing.T) {
	t.Parallel()

	kindGroup := funcMap["kindGroup"].(func(string) string)

	assert.Equal(t, "state", kindGroup(KindStruct))
	assert.Equal(t, "state", kindGroup(KindMap))
	assert.Equal(t, "state", kindGroup(KindSlice))
	assert.Equal(t, "state", kindGroup(KindArray))
	assert.Equal(t, "state", kindGroup(KindPointer))
	assert.Equal(t, "state", kindGroup(KindRef))
	assert.Equal(t, "code", kindGroup(KindFunc))
	assert.Equal(t, "code", kindGroup(KindClosure))
	assert.Equal(t, "types", kindGroup(KindType))
	assert.Equal(t, "types", kindGroup(KindInterface))
	assert.Equal(t, "other", kindGroup(KindPrimitive))
	assert.Equal(t, "other", kindGroup("unknown"))
}

func TestTemplateFunc_QueryHas(t *testing.T) {
	t.Parallel()

	queryHas := funcMap["queryHas"].(func(url.Values, string) bool)

	assert.False(t, queryHas(nil, "any"), "nil values should return false")
	assert.False(t, queryHas(url.Values{}, "missing"))
	assert.True(t, queryHas(url.Values{"state": {""}}, "state"))
}

func TestTemplateFunc_Dict(t *testing.T) {
	t.Parallel()

	dict := funcMap["dict"].(func(...any) (map[string]any, error))

	out, err := dict("a", 1, "b", "two")
	require.NoError(t, err)
	assert.Equal(t, 1, out["a"])
	assert.Equal(t, "two", out["b"])

	_, err = dict("a", 1, "b")
	assert.Error(t, err, "odd number of args should error")

	_, err = dict(1, "value")
	assert.Error(t, err, "non-string key should error")
}
