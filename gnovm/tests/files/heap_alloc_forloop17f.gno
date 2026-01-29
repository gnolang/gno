package main


var p *int

func main() {
	for i := 0; i < 10; i++ {
		switch i % 3 {
		case 0:
			continue // Targets the for loop
		case 1:
			println("One")
			continue // Targets the for loop
		}
		
		p = &i
	}
	println(*p)
}

// Preprocessed:
// file{ package main; var p<!~VPBlock(2,0)> *((const-type int)); func main() { for .loopvar_i<!~VPBlock(1,0)> := (const (0 int)); .loopvar_i<~VPBlock(1,0)> < (const (10 int)); .loopvar_i<~VPBlock(1,0)>++ { switch .loopvar_i<~VPBlock(2,0)> % (const (3 int)) { case (const (0 int)): .loopvar_i<@!~VPBlock(2,0)> = .loopvar_i<~VPBlock(2,0)>; continue; case (const (1 int)): (const (println func(...interface {})))((const ("One" string))); .loopvar_i<@!~VPBlock(2,0)> = .loopvar_i<~VPBlock(2,0)>; continue }; p<~VPBlock(4,0)> = &(.loopvar_i<~VPBlock(1,0)>); .loopvar_i<@!~VPBlock(1,0)> = .loopvar_i<~VPBlock(1,0)> }; (const (println func(...interface {})))(*(p<~VPBlock(3,0)>)) } }

// Output:
// One
// One
// One
// 8
