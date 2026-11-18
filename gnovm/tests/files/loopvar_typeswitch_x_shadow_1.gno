package main

// A type switch whose var shadows an enclosing for-loop var, while the
// switch expression X reads that enclosing loop var. X is evaluated in the
// enclosing scope, so its x must be renamed to the loop's x.loopvar; only
// the clauses see the type switch's own x. Previously the var was dropped
// from the rename scope before X was walked, and preprocess failed with
// "name x not declared".
func main() {
	for x := interface{}(1); x != nil; x = nil {
		switch x := x.(type) {
		case int:
			println("int", x)
		default:
			println("other")
		}
	}
}

// Preprocessed:
// file{ package main; func main() { for x.loopvar<!VPBlock(1,0)> := (const (1 int)); x.loopvar<VPBlock(1,0)> != (const (undefined)); x.loopvar<VPBlock(1,0)> = (const (undefined)) { switch x:=x.loopvar<VPBlock(2,0)> { case (const-type int): (const (println func(...interface {})))((const ("int" string)), x<VPBlock(1,0)>); default: (const (println func(...interface {})))((const ("other" string))) } } } }

// Output:
// int 1
