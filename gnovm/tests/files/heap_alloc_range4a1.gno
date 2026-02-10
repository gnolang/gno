package main

func main() {
	var fns []func() int
	m := map[string]int{"a": 1, "b": 2}
	for _, v := range m {
		f := func() int {
			return v
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func() .res.0 (const-type int); m<!VPBlock(1,1)> := (const-type map[string]int){(const ("a" string)): (const (1 int)), (const ("b" string)): (const (2 int))}; for _<VPBlock(0,0)>, v<!~VPBlock(1,0)> := range m<VPBlock(2,1)> { f<!VPBlock(1,1)> := func func() .res.0 (const-type int){ return v<~VPBlock(1,1)> }<v<()~VPBlock(1,0)>>; fns<VPBlock(2,0)> = (const (append func([]func() int, ...func() int) []func() int))(fns<VPBlock(2,0)>, f<VPBlock(1,1)>) }; for _<VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(...interface {})))(fn<VPBlock(1,0)>()) } } }

// Output:
// 1
// 2
