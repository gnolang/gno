package main

type rand struct{}

func (r *rand) Seed() {
	println("seed...")
}

func fromSeed() *rand {
	return &rand{}
}

// XXX, test order in correcit with defer
// [params, res_xxx, vars]
func genResult(s0 string, x int) (int, bool) {
	z := 0
	println(z)
	r := fromSeed()
	defer func() { r.Seed() }()

	return -1, true
}

func main() {
	genResult("hey", 0)
}

// Output:
// 0
// seed...
