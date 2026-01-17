package rules

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestHasEmptyStringBounds(t *testing.T) {
	emptyStr := &gnolang.BasicLitExpr{Kind: gnolang.STRING, Value: `""`}
	nonEmptyStr := &gnolang.BasicLitExpr{Kind: gnolang.STRING, Value: `"foo"`}

	tests := []struct {
		name string
		args []gnolang.Expr
		want bool
	}{
		{"both empty", []gnolang.Expr{emptyStr, emptyStr}, true},
		{"first non-empty", []gnolang.Expr{nonEmptyStr, emptyStr}, false},
		{"second non-empty", []gnolang.Expr{emptyStr, nonEmptyStr}, false},
		{"both non-empty", []gnolang.Expr{nonEmptyStr, nonEmptyStr}, false},
		{"no args", []gnolang.Expr{}, false},
		{"one arg", []gnolang.Expr{emptyStr}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &gnolang.CallExpr{Args: tt.args}
			if got := hasEmptyStringBounds(call); got != tt.want {
				t.Errorf("hasEmptyStringBounds() = %v, want %v", got, tt.want)
			}
		})
	}
}
