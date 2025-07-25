package main

func main() {
	var i any = "truc"

	switch b := 2; a := i.(type) {
	case string:
		println("string", a+" ok")
	default:
		println("unknown", b)
	}
}

// Preprocessed:
// file{ package main; func main() { var i<!VPBlock(1,0)> (const-type interface {}) = (const ("truc" string)); switch b<!VPBlock(1,0)> := (const (2 int)); a:=i<VPBlock(2,0)> { case (const-type string): (const (println func(...interface {})))((const ("string" string)), a<VPBlock(1,1)> + (const (" ok" string))); default: (const (println func(...interface {})))((const ("unknown" string)), b<VPBlock(1,0)>) } } }

// Output:
// string truc ok
