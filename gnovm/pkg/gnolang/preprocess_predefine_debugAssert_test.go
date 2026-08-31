//go:build debugAssert

package gnolang

import (
	"strings"
	"testing"
)

func TestPredefineDeclIndexRejectsReplacedDecl(t *testing.T) {
	decl := &ValueDecl{NameExprs: NameExprs{{Name: "value"}}}
	file := &FileNode{Decls: Decls{decl}}
	index := newPredefineDeclIndex(&FileSet{Files: []*FileNode{file}})
	file.Decls[0] = &ValueDecl{NameExprs: NameExprs{{Name: "value"}}}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("lookup accepted a replaced declaration")
		}
		if msg, _ := r.(string); !strings.Contains(msg, "stale predefine declaration index") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	index.lookup("value")
}
