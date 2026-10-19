// Copyright 2026 The Gno Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"fmt"
	"go/ast"
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
		{
			name: "follow-up errors at EOF are dropped",
			src: `package p

func f() {
	x := f(g(
`,
			want: []string{
				"4:12: expected ')', found 'EOF'",
			},
		},
		{
			name: "missing comma before a newline ends the list", // issue 22164
			src: `package p

func f() {
	x := f(g()
	y := 1
}

func g() {
	x := []int{1, 2, 3
	y := 0
}
`,
			want: []string{
				"4:12: unexpected newline in argument list; possibly missing comma or )",
				"9:20: unexpected newline in composite literal; possibly missing comma or }",
			},
		},
		{
			name: "missing comma before a newline in a parameter list",
			src: `package p

func h(a int
	b int) {
}

func i[T any
	U any](x T) {
}
`,
			want: []string{
				"3:13: unexpected newline in parameter list; possibly missing comma or )",
				"7:13: unexpected newline in type parameter list; possibly missing comma or ]",
			},
		},
		{
			name: "label without a statement", // bug274, issue 14006
			src: `package p

func _() {
	switch 0 {
	case 0:
		L0:
	case 1:
		L1:
	default:
		L2:
	}
	2:
	L3:
}
`,
			want: []string{
				"6:5: missing statement after label",
				"8:5: missing statement after label",
				"12:3: illegal label declaration",
			},
		},
		{
			name: "statements at file level are each reported", // syntax/topexpr
			src: `package p

fmt.Printf("hello")

func main() {
}

x++

func init() {
}

x, y := 1, 2
`,
			want: []string{
				"3:1: non-declaration statement outside function body",
				"8:1: non-declaration statement outside function body",
				"13:1: non-declaration statement outside function body",
			},
		},
		{
			name: "statement header resynchronises at its block", // issue 22581
			src: `package p

func f() {
	if i := g()); i == j {
	}
	if i == g()] {
	}
	switch i := g()); i {
	}
	switch g()] {
	}
	for i := g()); i < y; {
	}
	for g()] {
	}
	switch x; y
	{
		z
`,
			want: []string{
				"4:13: expected ';', found ')'",
				"6:13: expected ';', found ']'",
				"8:9: expected switch expression, found assignment (missing parentheses around composite literal?)",
				"8:17: expected '{', found ')'",
				"10:12: expected '{', found ']'",
				"12:6: expected boolean or range expression, found assignment (missing parentheses around composite literal?)",
				"12:14: expected '{', found ')'",
				"14:9: expected '{', found ']'",
				"16:13: expected '{', found newline",
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

// TestRecoveryLabelTree pins the tree a label with no statement leaves:
// an implicit empty statement placed right after its colon, so the
// label's own line is the statement's.
func TestRecoveryLabelTree(t *testing.T) {
	fset := token.NewFileSet()
	f, _ := ParseFile(fset, "y.go", "package p\n\nfunc _() {\n\tswitch {\n\tcase true:\n\t\tL:\n\t}\n\tM:\n}\n", AllErrors)
	ast.Inspect(f, func(n ast.Node) bool {
		if l, ok := n.(*ast.LabeledStmt); ok {
			e, ok := l.Stmt.(*ast.EmptyStmt)
			if !ok || !e.Implicit || e.Semicolon != l.Colon+1 {
				t.Errorf("label %s: statement %#v, want an implicit empty statement after the colon", l.Label.Name, l.Stmt)
			}
		}
		return true
	})
}

// TestRecoveryDeclTree pins that the declarations after a file-level
// statement are kept.
func TestRecoveryDeclTree(t *testing.T) {
	fset := token.NewFileSet()
	f, _ := ParseFile(fset, "x.go", "package p\n\nx++\n\nfunc main() {}\n\nx++\n\nfunc init() {}\n", AllErrors)
	var kinds []string
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			kinds = append(kinds, "func "+d.Name.Name)
		default:
			kinds = append(kinds, fmt.Sprintf("%T", d))
		}
	}
	if got, want := strings.Join(kinds, ", "), "*ast.BadDecl, func main, *ast.BadDecl, func init"; got != want {
		t.Errorf("decls: %s, want %s", got, want)
	}
}
