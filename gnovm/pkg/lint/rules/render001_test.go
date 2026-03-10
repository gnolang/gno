package rules

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/lint"
)

func TestRENDER001_Info(t *testing.T) {
	rule := RENDER001{}
	info := rule.Info()

	if info.ID != "RENDER001" {
		t.Errorf("expected ID RENDER001, got %s", info.ID)
	}
	if info.Severity != lint.SeverityError {
		t.Errorf("expected SeverityError, got %s", info.Severity)
	}
}

func TestRENDER001_NonFuncDecl(t *testing.T) {
	rule := RENDER001{}
	ctx := &lint.RuleContext{
		PkgPath: "gno.land/r/test/myrealm",
		File:    &gnolang.FileNode{FileName: "test.gno"},
	}

	// Non-FuncDecl nodes should return nil.
	issues := rule.Check(ctx, &gnolang.ValueDecl{})
	if issues != nil {
		t.Errorf("expected nil for non-FuncDecl, got %v", issues)
	}
}

func TestRENDER001_NotRender(t *testing.T) {
	rule := RENDER001{}
	ctx := &lint.RuleContext{
		PkgPath: "gno.land/r/test/myrealm",
		File:    &gnolang.FileNode{FileName: "test.gno"},
		Parents: []gnolang.Node{&gnolang.FileNode{}},
	}

	fn := &gnolang.FuncDecl{}
	fn.Name = "NotRender"
	issues := rule.Check(ctx, fn)
	if issues != nil {
		t.Errorf("expected nil for non-Render function, got %v", issues)
	}
}

func TestRENDER001_Method(t *testing.T) {
	rule := RENDER001{}
	ctx := &lint.RuleContext{
		PkgPath: "gno.land/r/test/myrealm",
		File:    &gnolang.FileNode{FileName: "test.gno"},
		Parents: []gnolang.Node{&gnolang.FileNode{}},
	}

	fn := &gnolang.FuncDecl{IsMethod: true}
	fn.Name = "Render"
	issues := rule.Check(ctx, fn)
	if issues != nil {
		t.Errorf("expected nil for Render method, got %v", issues)
	}
}

func TestRENDER001_NotRealm(t *testing.T) {
	rule := RENDER001{}
	ctx := &lint.RuleContext{
		PkgPath: "gno.land/p/test/mypkg",
		File:    &gnolang.FileNode{FileName: "test.gno"},
		Parents: []gnolang.Node{&gnolang.FileNode{}},
	}

	fn := &gnolang.FuncDecl{}
	fn.Name = "Render"
	issues := rule.Check(ctx, fn)
	if issues != nil {
		t.Errorf("expected nil for non-realm package, got %v", issues)
	}
}

func TestRENDER001_NotTopLevel(t *testing.T) {
	rule := RENDER001{}
	ctx := &lint.RuleContext{
		PkgPath: "gno.land/r/test/myrealm",
		File:    &gnolang.FileNode{FileName: "test.gno"},
		Parents: []gnolang.Node{&gnolang.FuncDecl{}}, // parent is another func, not file
	}

	fn := &gnolang.FuncDecl{}
	fn.Name = "Render"
	issues := rule.Check(ctx, fn)
	if issues != nil {
		t.Errorf("expected nil for non-top-level Render, got %v", issues)
	}
}
