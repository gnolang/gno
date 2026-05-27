package main

// Tests that dependencies are correctly discovered through doubly-nested
// function literals. The NameExpr for B is two FuncLitExprs deep;
// addDependencyToTopDecl must walk past both to find ValueDecl(A).

var A = func() int {
	return func() int { return B }()
}()

var B = 42

func main() { println(A) }

// Output:
// 42
