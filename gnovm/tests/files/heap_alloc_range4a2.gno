package main

func main() {
	var fns []func() int
	y := 0
	m := map[string]int{"a": 1, "b": 2}
	for _, v := range m {
		x := v
		f := func() int {
			switch y {
			case 0:
				if true {
					return x
				}
			default:
				return 0
			}
			return 0
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Preprocessed:
// file{ package main; func main() { var fns<!VPBlock(1,0)> []func()  (const-type int); y<!VPBlock(1,1)> := (const (0 int)); m<!VPBlock(1,2)> := map[(const-type string)] (const-type int){(const ("a" string)): (const (1 int)), (const ("b" string)): (const (2 int))}; for _<!~VPBlock(0,0)>, v<!VPBlock(1,0)> := range m<VPBlock(2,2)> { x<!~VPBlock(1,1)> := v<VPBlock(1,0)>; f<!VPBlock(1,2)> := func func()  (const-type int){ switch y<VPBlock(4,1)> { case (const (0 int)): if (const (true bool)) { return x<~VPBlock(3,1)> }; default: return (const (0 int)) }; return (const (0 int)) }<x<()~VPBlock(1,1)>>; fns<VPBlock(2,0)> = (const (append func(x []func()( int),args ...func()( int))(res []func()( int))))(fns<VPBlock(2,0)>, f<VPBlock(1,2)>) }; for _<!~VPBlock(0,0)>, fn<!VPBlock(1,0)> := range fns<VPBlock(2,0)> { (const (println func(xs ...interface{})()))(fn<VPBlock(1,0)>()) } } }

// Output:
// 1
// 2
