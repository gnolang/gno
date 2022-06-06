package ufmt

import (
	"fmt"
	"testing"
)

func TestSprintf(t *testing.T) {
	cases := []struct {
		format        string
		values        []interface{}
		exectedOutput string
	}{
		{"hello %s!", []interface{}{"planet"}, "hello planet!"},
		{"hi %%%s!", []interface{}{"worl%d"}, "hi %worl%d!"},
	}

	for _, tc := range cases {
		name := fmt.Sprintf(tc.format, tc.values...)
		println(name)
		t.Run(name, func(t *testing.T) {
			t.Log("blah")
		})
	}
}
