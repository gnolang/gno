package nameregistry_test

import (
	"strings"
	"testing"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/uassert"

	"gno.land/r/sys/nameregistry"
)

func TestGenHomoglyphVariants(t *testing.T) {
	tests := []struct {
		name     string
		basename string
		variants string
	}{
		{
			name:     "no variants",
			basename: "foobar",
		},
		{
			name:     "one homoglyph",
			basename: "f0obar",
			variants: "foobar",
		},
		{
			name:     "two homoglyphs",
			basename: "f00bar",
			variants: "fo0bar,f0obar,foobar",
		},
		{
			name:     "mixed homoglyphs",
			basename: "0xlx1",
			variants: "oxlx1,0xix1,oxix1,0xlxi,oxlxi,0xixi,oxixi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			variants := nameregistry.GenHomoglyphVariants(tt.basename)

			// Assert
			uassert.Equal(t, tt.variants, strings.Join(variants, ","), "expect variants to match")
		})
	}
}

func TestCheckNameSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		basename string
		names    []string
		errMsg   string
	}{
		{
			name:     "ok",
			basename: "foo",
			names:    []string{"bar"},
		},
		{
			name:     "reserved name",
			basename: "foo",
			names:    []string{"foo", "bar"},
			errMsg:   "name is reserved",
		},
		{
			name:     "reserved name with typosquatting",
			basename: "f00",
			names:    []string{"foo", "bar"},
			errMsg:   "similar name already exists: foo",
		},
		{
			name:     "reserved name with numbers suffix",
			basename: "foo000",
			names:    []string{"foo", "bar"},
			errMsg:   "similar name already exists: foo",
		},
		{
			name:     "reserved name with numbers and letter suffix",
			basename: "foo000bar000",
			names:    []string{"foo", "bar"},
			errMsg:   "similar name already exists: foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			names := avl.NewTree()
			for _, n := range tt.names {
				names.Set(n, struct{}{})
			}

			// Act
			err := nameregistry.CheckNameSimilarity(tt.basename, names)

			// Assert
			if tt.errMsg != "" {
				uassert.ErrorContains(t, err, tt.errMsg, "expect an error")
			} else {
				uassert.Nil(t, err, "expect no error")
			}
		})
	}
}
