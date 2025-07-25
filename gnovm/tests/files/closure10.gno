package main

func main() {
	var x any = int(1)
	if _, ok := x.(string); !ok { // !ok
		x = func() bool {
			return ok
		}
	}
	ok2 := x.(func() bool)()
	println(ok2)
}

// Preprocessed:
// file{ package main; func main() { var x<!VPBlock(1,0)> (const-type interface {}) = (const (1 int)); if _<VPBlock(0,0)>, ok<!~VPBlock(1,0)> := x<VPBlock(2,0)>.((const-type string)); !ok<~VPBlock(1,0)> { x<VPBlock(2,0)> = func func() .res.0 (const-type bool){ return ok<~VPBlock(1,1)> }<ok<()~VPBlock(1,0)>> }; ok2<!VPBlock(1,1)> := x<VPBlock(1,0)>.(func() .res.0 (const-type bool))(); (const (println func(...interface {})))(ok2<VPBlock(1,1)>) } }

// Output:
// false
