package main

func main() {
	s := []int{1, 2}

	f := func() {
		for i, v := range s {
			println(i)
			println(v)
		}
	}

	f()
}

// Preprocessed:
// file{ package main; func main() { s<!~VPBlock(1,0)> := (const-type []int){(const (1 int)), (const (2 int))}; f<!VPBlock(1,1)> := func func(){ for i.loopvar<!VPBlock(1,0)>, v.loopvar<!VPBlock(1,1)> := range s<~VPBlock(2,0)> { (const (println func(...interface {})))(i.loopvar<VPBlock(1,0)>); (const (println func(...interface {})))(v.loopvar<VPBlock(1,1)>) } }<s<()~VPBlock(1,0)>>; f<VPBlock(1,1)>() } }

// Output:
// 0
// 1
// 1
// 2
