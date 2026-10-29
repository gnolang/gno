package gnolang

// Checker walks exposed to check_gas_bench_test.go (package gnolang_test,
// which needs pkg/test's example store: an import cycle from here).
var (
	CheckUnusedForTest              = checkUnused
	CheckUnusedImportsForTest       = checkUnusedImports
	CheckInitCyclesForTest          = checkInitCycles
	CheckTypeShapeForTest           = checkTypeShape
	CheckTypeExprShapeForTest       = checkTypeExprShape
	AssertNotImportedForTest        = assertNotImported
	DuplicateSwitchCasesForTest     = duplicateSwitchCases
	DuplicateTypeSwitchCasesForTest = duplicateTypeSwitchCases
)

// DeclaredTypeOfForTest returns the type a preprocessed non-alias TypeDecl
// declared and the type expression it was evaluated from; nil, nil for
// an alias or an unpreprocessed declaration.
func DeclaredTypeOfForTest(td *TypeDecl) (*DeclaredType, Expr) {
	ctx, ok := td.Type.(*constTypeExpr)
	if !ok || td.IsAlias {
		return nil, nil
	}
	dt, _ := ctx.Type.(*DeclaredType)
	return dt, ctx.Source
}
