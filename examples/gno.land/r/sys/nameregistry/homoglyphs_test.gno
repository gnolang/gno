package nameregistry_test

import (
	"strings"
	"testing"

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
