package gnolang

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCheckNoGenerics(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name, body, wantErr string
	}{
		{
			"GenericTypeDecl",
			`package hello
			type empty interface{}
			type Foo[T empty] int`,
			"generic type declarations are not supported",
		},
		{
			"GenericFunc",
			`package hello
			func Bar[T any]() {}`,
			"generic functions are not supported",
		},
		{
			"TypeUnion",
			`package hello
			type N interface{ int | string }`,
			"interface type unions are not supported",
		},
		{
			"Approximation",
			`package hello
			type N interface{ ~int }`,
			"interface approximation (~) terms are not supported",
		},
		{
			"PlainInterfaceEmbedOK",
			`package hello
			type A interface{ M() }
			type B interface{ A }
			var _ B`,
			"",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mpkg := &std.MemPackage{
				Type: MPUserProd,
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*std.MemFile{
					{Name: "hello.gno", Body: tc.body},
				},
			}
			_, err := TypeCheckMemPackage(mpkg, TypeCheckOptions{Mode: TCLatestRelaxed})
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected accept, got: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got: %v", tc.wantErr, err)
			}
		})
	}
}
