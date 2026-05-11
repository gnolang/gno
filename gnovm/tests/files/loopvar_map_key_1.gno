package main

// Map composite literal: the loop var name is used as a map key expression.
// Unlike struct fields, map keys ARE variable references and MUST be renamed.
// This also exercises the TRANS_COMPOSITE_KEY fix for struct field names.
func main() {
	for k := range []string{"a", "b", "c"} {
		m := map[int]string{k: "val"}
		println(k, m[k])
	}
}

// Output:
// 0 val
// 1 val
// 2 val
