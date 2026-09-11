// Regression: ASSIGN-form range (`= range`, not `:=`) with a non-NameExpr LHS
// (IndexExpr) used to read the range subject at a fixed value-stack offset that
// only held for NameExpr targets, so it ranged over an LHS pointer operand
// instead of X. Covers slice, array, array-pointer, string, and map subjects,
// and the key-only form (`for m[k] = range x`, where bs.Value is nil).
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

	// array subject, IndexExpr key and value targets.
	arr := [3]string{"A", "B", "C"}
	ak := map[string]int{"k": -1}
	av := map[string]string{"k": "?"}
	for ak["k"], av["k"] = range arr {
	}
	println(ak["k"], av["k"], len(av["k"]))

	// array-pointer subject: the only handler that writes back through the
	// subject slot (*xv = *dv), so a wrong index corrupts it rather than
	// only misreading it.
	pa := &arr
	pk := map[string]int{"k": -1}
	pv := map[string]string{"k": "?"}
	for pk["k"], pv["k"] = range pa {
	}
	println(pk["k"], pv["k"], len(pv["k"]))

	// key-only form: bs.Value is nil, so the operands above X come from the
	// key target alone.
	ko := map[string]int{"k": -1}
	for ko["k"] = range []string{"A", "B", "C", "D"} {
	}
	println(ko["k"])

	ko2 := map[string]int{"k": -1}
	for ko2["k"] = range "XYZ" {
	}
	println(ko2["k"])

	ko3 := map[string]int{"k": -1}
	for ko3["k"] = range src {
	}
	println(ko3["k"])
}

// Output:
// 4 E 1
// 1 89
// 7 hello
// 2 C 1
// 2 C 1
// 3
// 2
// 7
