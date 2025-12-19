package main

var ps []*int

func main() {
	for i := 0; i < 3; i++ {
		ps = append(ps, &i)
	}
	for _, v := range ps{
		println(*v)
	}
}

// Preprocessed:
// file{ package main; var ps<!~VPBlock(2,0)> []*((const-type int)); func main() { for .loopvar_i<!@VPBlock(1,0)> := (const (0 int)); .loopvar_i<*VPBlock(1,0)> < (const (3 int)); .loopvar_i<*VPBlock(1,0)>++ { .redefine_i<!~VPBlock(1,0)> := .loopvar_i<*VPBlock(2,0)>; ps<~VPBlock(5,0)> = (const (append func([]*int, ...*int) []*int))(ps<~VPBlock(5,0)>, &(.redefine_i<~VPBlock(1,0)>)); .loopvar_i<*VPBlock(2,0)> = .redefine_i<~VPBlock(1,0)> } ; for _<VPBlock(0,0)>, v<!VPBlock(1,0)> := range ps<~VPBlock(4,0)> { (const (println func(...interface {})))(*(v<VPBlock(1,0)>)) } } }

// Output:
// 0
// 1
// 2
