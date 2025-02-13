package gnolang

import (
	"fmt"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseForLoop(t *testing.T) {
	t.Parallel()

	gocode := `package main
func main(){
	for i:=0; i<10; i++ {
		if i == -1 {
			return
		}
	}
}`
	n, err := ParseFile("main.go", gocode)
	assert.NoError(t, err, "ParseFile error")
	assert.NotNil(t, n, "ParseFile error")
	fmt.Printf("CODE:\n%s\n\n", gocode)
	fmt.Printf("AST:\n%#v\n\n", n)
	fmt.Printf("AST.String():\n%s\n", n.String())
}

func TestLocationPlusError(t *testing.T) {
	recoverAsError := func(fn func()) (err error) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			if re, ok := r.(error); ok {
				if re != nil {
					err = re
				}
			} else {
				err = fmt.Errorf("%v", r)
			}
		}()

		fn()
		return
	}

	tests := []struct {
		name    string
		fn      func()
		wantErr string
	}{
		{
			"with well defined position and msg",
			func() {
				panic(&LocationPlusError{
					pos: token.Position{
						Filename: "a.gno",
						Line:     10, Column: 33,
					},
					msg: "here",
				})
			},
			"a.gno:10:33: here",
		},
		{
			"with blank msg",
			func() {
				panic(&LocationPlusError{
					pos: token.Position{
						Filename: "a.gno",
						Line:     10, Column: 33,
					},
					msg: "",
				})
			},
			"a.gno:10:33: ",
		},
		{
			"with undefined Line and Column",
			func() {
				panic(&LocationPlusError{
					pos: token.Position{
						Filename: "a.gno",
						Line:     0, Column: 0,
					},
					msg: "",
				})
			},
			"a.gno:0:0: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := recoverAsError(tt.fn)
			if tt.wantErr == "" {
				require.NoError(t, got)
			}

			require.Contains(t, got.Error(), tt.wantErr)
		})
	}
}
