package main

// A range whose key/value shadow an enclosing for-loop var, while the range
// expression X reads that enclosing loop var. X is evaluated in the
// enclosing scope, so its x must be renamed to the loop's x.loopvar; only
// the body sees the range's own x. Previously the key/value were dropped
// from the rename scope before X was walked, and preprocess failed with
// "name x not declared".
func main() {
	for x := 1; x < 3; x++ {
		for x := range make([]int, x) {
			println(x)
		}
	}
	for x := 5; x < 6; x++ {
		for _, x := range []int{x, x + 1} {
			println(x)
		}
	}
	// X is the bare loop var itself.
	for xs := []int{9}; len(xs) < 2; xs = append(xs, 0) {
		for _, xs := range xs {
			println(xs)
		}
	}
}

// Preprocessed:
// file{ package main; func main() { for x.loopvar<!VPBlock(1,0)> := (const (1 int)); x.loopvar<VPBlock(1,0)> < (const (3 int)); x.loopvar<VPBlock(1,0)>++ { for x<!VPBlock(1,0)> := range (const (make func(type{}, ...interface {}) []int))([](const-type int), x.loopvar<VPBlock(2,0)>) { (const (println func(...interface {})))(x<VPBlock(1,0)>) } }; for x.loopvar<!VPBlock(1,0)> := (const (5 int)); x.loopvar<VPBlock(1,0)> < (const (6 int)); x.loopvar<VPBlock(1,0)>++ { for _<VPBlock(0,0)>, x<!VPBlock(1,0)> := range (const-type []int){x.loopvar<VPBlock(2,0)>, x.loopvar<VPBlock(2,0)> + (const (1 int))} { (const (println func(...interface {})))(x<VPBlock(1,0)>) } }; for xs.loopvar<!VPBlock(1,0)> := (const-type []int){(const (9 int))}; (const (len func(interface {}) int))(xs.loopvar<VPBlock(1,0)>) < (const (2 int)); xs.loopvar<VPBlock(1,0)> = (const (append func([]int, ...int) []int))(xs.loopvar<VPBlock(1,0)>, (const (0 int))) { for _<VPBlock(0,0)>, xs<!VPBlock(1,0)> := range xs.loopvar<VPBlock(2,0)> { (const (println func(...interface {})))(xs<VPBlock(1,0)>) } } } }

// Output:
// 0
// 0
// 1
// 5
// 6
// 9
