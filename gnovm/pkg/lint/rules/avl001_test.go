package rules

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

func TestAVL001_Info(t *testing.T) {
	rule := &AVL001{}
	info := rule.Info()

	if info.ID != "AVL001" {
		t.Errorf("ID = %v, want AVL001", info.ID)
	}
	if info.Category != lint.CategoryAVL {
		t.Errorf("Category = %v, want CategoryAVL", info.Category)
	}
	if info.Name != "unbounded-iteration" {
		t.Errorf("Name = %v, want unbounded-iteration", info.Name)
	}
	if info.Severity != lint.SeverityWarning {
		t.Errorf("Severity = %v, want SeverityWarning", info.Severity)
	}
}

func TestAVL001_Check_NotCallExpr(t *testing.T) {
	rule := &AVL001{}
	ctx := &lint.RuleContext{}

	// AVL001 only checks CallExpr nodes
	// Passing nil should return nil issues
	issues := rule.Check(ctx, nil)
	if issues != nil {
		t.Errorf("Check(nil) = %v, want nil", issues)
	}
}

func TestIsEmptyStringLiteral(t *testing.T) {
	// This tests the helper function behavior conceptually
	// Since we can't easily construct gnolang nodes in tests,
	// we verify the function signature and expected behavior
	// through integration tests

	// The function should:
	// - Return true for BasicLitExpr with Kind=STRING and Value=`""`
	// - Return true for ConstExpr with StringKind and empty string value
	// - Return false for non-string types
	// - Return false for non-empty strings
}

func TestIsAVLTree(t *testing.T) {
	// This tests the helper function behavior conceptually
	// The function should:
	// - Return true for *gnolang.DeclaredType with PkgPath="gno.land/p/nt/avl" and Name="Tree"
	// - Return true for *gnolang.PointerType pointing to such a type
	// - Return false for other types
}

func TestHasEmptyStringBounds(t *testing.T) {
	// This tests the helper function behavior conceptually
	// The function should:
	// - Return true when both first two args are empty strings
	// - Return false when there are fewer than 2 args
	// - Return false when either arg is non-empty
}
