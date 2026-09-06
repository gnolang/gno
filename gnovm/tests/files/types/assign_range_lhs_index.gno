// Regression: ASSIGN-form range (`= range`, not `:=`) with a non-NameExpr LHS
// (IndexExpr) used to read the range subject at a fixed value-stack offset that
// only held for NameExpr targets, so it ranged over an LHS pointer operand
// instead of X. Covers slice, string, and map subjects.
package main

func main() {
	// slice subject, IndexExpr value target.
	ss := map[string]string{"k": "orig"}
	var i int
	for i, ss["k"] = range []string{"A", "B", "C", "D", "E"} {
	}
	println(i, ss["k"], len(ss["k"]))

	// string subject, IndexExpr value target (rune).
	rr := map[string]rune{"k": 'z'}
	var j int
	for j, rr["k"] = range "XY" {
	}
	println(j, int(rr["k"]))

	// map subject, IndexExpr key and value targets.
	src := map[int]string{7: "hello"}
	kk := map[int]int{0: -1}
	vv := map[int]string{0: "?"}
	for kk[0], vv[0] = range src {
	}
	println(kk[0], vv[0])
}

// Output:
// 4 E 1
// 1 89
// 7 hello
