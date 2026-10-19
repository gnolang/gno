// Copyright 2026 The Gno Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"fmt"
	"go/scanner"
	"go/token"
	"strings"
	"testing"
)

// TestRecovery pins the error positions and the resynchronisation the
// parser shares with gc (cmd/compile/internal/syntax): one error per
// mistake, no follow-up on the lines after it, and in AllErrors mode,
// since that is how Gno's accumulate mode runs it.
func TestRecovery(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string // "line:col: msg", in position order
	}{
		{
			name: "missing operand stops at the closing token", // issue 33386
			src: `package p

func _() {
	defer func() {
		1 +
	}()
}

func _() {
	_ = (1 +)
	_ = a[2 +]
	_ = []int{1, 2, 3 + }
}
`,
			want: []string{
				"6:2: expected operand, found '}'",
				"10:10: expected operand, found ')'",
				"11:11: expected operand, found ']'",
				"12:22: expected operand, found '}'",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			_, err := ParseFile(fset, "x.go", tt.src, AllErrors)
			list, _ := err.(scanner.ErrorList)
			var got []string
			for _, e := range list {
				got = append(got, fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Column, e.Msg))
			}
			if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
				t.Errorf("errors:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}
