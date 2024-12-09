package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexMap(t *testing.T) {
	t.Parallel()

	m := map[string]any{
		"Key1": map[string]any{
			"Key2": map[string]any{
				"Key3": 1,
			},
		},
	}

	test := IndexMap(m)
	assert.NotNil(t, test, "should return m")
	_, ok := test.(map[string]any)
	assert.True(t, ok, "returned m should be a map")

	test = IndexMap(m, "Key1")
	assert.NotNil(t, test, "should return Key1 value")
	_, ok = test.(map[string]any)
	assert.True(t, ok, "Key1 value type should be a map")

	test = IndexMap(m, "Key1", "Key2")
	assert.NotNil(t, test, "should return Key2 value")
	_, ok = test.(map[string]any)
	assert.True(t, ok, "Key2 value type should be a map")

	test = IndexMap(m, "Key1", "Key2", "Key3")
	assert.NotNil(t, test, "should return Key3 value")
	val, ok := test.(int)
	assert.True(t, ok, "Key3 value type should be an int")
	assert.Equal(t, 1, val, "Key3 value should be a 1")

	test = IndexMap(m, "Key1", "Key2", "Key3", "Key4")
	assert.Nil(t, test, "Key4 value should not exist")
}
