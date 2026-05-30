// Ported from upstream Go's sort examples at go1.25.9.
// Combines content from upstream example_test.go, example_search_test.go,
// example_keys_test.go, example_multi_test.go, and example_wrapper_test.go.
// Skipped:
//   - ExampleFind: sort.Find not in Gno
//   - ExampleSlice/SliceStable/SliceIsSorted: depend on sort.Slice (not in Gno)
//   - Example: uses sort.Slice in its second half

package sort_test

import (
	"fmt"
	"math"
	"sort"
)

// --- example_test.go ---

func ExampleInts() {
	s := []int{5, 2, 6, 3, 1, 4} // unsorted
	sort.Ints(s)
	fmt.Println(s)
	// Output: [1 2 3 4 5 6]
}

func ExampleIntsAreSorted() {
	s := []int{1, 2, 3, 4, 5, 6} // sorted ascending
	fmt.Println(sort.IntsAreSorted(s))

	s = []int{6, 5, 4, 3, 2, 1} // sorted descending
	fmt.Println(sort.IntsAreSorted(s))

	s = []int{3, 2, 4, 1, 5} // unsorted
	fmt.Println(sort.IntsAreSorted(s))

	// Output: true
	// false
	// false
}

func ExampleFloat64s() {
	s := []float64{5.2, -1.3, 0.7, -3.8, 2.6} // unsorted
	sort.Float64s(s)
	fmt.Println(s)

	s = []float64{math.Inf(1), math.NaN(), math.Inf(-1), 0.0} // unsorted
	sort.Float64s(s)
	fmt.Println(s)

	// Output: [-3.8 -1.3 0.7 2.6 5.2]
	// [NaN -Inf 0 +Inf]
}

func ExampleFloat64sAreSorted() {
	s := []float64{0.7, 1.3, 2.6, 3.8, 5.2} // sorted ascending
	fmt.Println(sort.Float64sAreSorted(s))

	s = []float64{5.2, 3.8, 2.6, 1.3, 0.7} // sorted descending
	fmt.Println(sort.Float64sAreSorted(s))

	s = []float64{5.2, 1.3, 0.7, 3.8, 2.6} // unsorted
	fmt.Println(sort.Float64sAreSorted(s))

	// Output: true
	// false
	// false
}

func ExampleReverse() {
	s := []int{5, 2, 6, 3, 1, 4} // unsorted
	sort.Sort(sort.Reverse(sort.IntSlice(s)))
	fmt.Println(s)
	// Output: [6 5 4 3 2 1]
}

func ExampleStrings() {
	s := []string{"Go", "Bravo", "Gopher", "Alpha", "Grin", "Delta"}
	sort.Strings(s)
	fmt.Println(s)
	// Output: [Alpha Bravo Delta Go Gopher Grin]
}

// --- example_search_test.go ---

// This example demonstrates searching a list sorted in ascending order.
func ExampleSearch() {
	a := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	x := 6

	i := sort.Search(len(a), func(i int) bool { return a[i] >= x })
	if i < len(a) && a[i] == x {
		fmt.Printf("found %d at index %d in %v\n", x, i, a)
	} else {
		fmt.Printf("%d not found in %v\n", x, a)
	}
	// Output:
	// found 6 at index 2 in [1 3 6 10 15 21 28 36 45 55]
}

// This example demonstrates searching a list sorted in descending order.
// The approach is the same as searching a list in ascending order,
// but with the condition inverted.
func ExampleSearch_descendingOrder() {
	a := []int{55, 45, 36, 28, 21, 15, 10, 6, 3, 1}
	x := 6

	i := sort.Search(len(a), func(i int) bool { return a[i] <= x })
	if i < len(a) && a[i] == x {
		fmt.Printf("found %d at index %d in %v\n", x, i, a)
	} else {
		fmt.Printf("%d not found in %v\n", x, a)
	}
	// Output:
	// found 6 at index 7 in [55 45 36 28 21 15 10 6 3 1]
}

func ExampleSearchFloat64s() {
	a := []float64{1.0, 2.0, 3.3, 4.6, 6.1, 7.2, 8.0}

	x := 2.0
	i := sort.SearchFloat64s(a, x)
	fmt.Printf("found %g at index %d in %v\n", x, i, a)

	x = 0.5
	i = sort.SearchFloat64s(a, x)
	fmt.Printf("%g not found, can be inserted at index %d in %v\n", x, i, a)
	// Output:
	// found 2 at index 1 in [1 2 3.3 4.6 6.1 7.2 8]
	// 0.5 not found, can be inserted at index 0 in [1 2 3.3 4.6 6.1 7.2 8]
}

func ExampleSearchInts() {
	a := []int{1, 2, 3, 4, 6, 7, 8}

	x := 2
	i := sort.SearchInts(a, x)
	fmt.Printf("found %d at index %d in %v\n", x, i, a)

	x = 5
	i = sort.SearchInts(a, x)
	fmt.Printf("%d not found, can be inserted at index %d in %v\n", x, i, a)
	// Output:
	// found 2 at index 1 in [1 2 3 4 6 7 8]
	// 5 not found, can be inserted at index 4 in [1 2 3 4 6 7 8]
}

func ExampleSearchStrings() {
	a := []string{"apple", "banana", "cherry", "date", "fig", "grape"}

	x := "banana"
	i := sort.SearchStrings(a, x)
	fmt.Printf("found %s at index %d in %v\n", x, i, a)

	x = "coconut"
	i = sort.SearchStrings(a, x)
	fmt.Printf("%s not found, can be inserted at index %d in %v\n", x, i, a)

	// Output:
	// found banana at index 1 in [apple banana cherry date fig grape]
	// coconut not found, can be inserted at index 3 in [apple banana cherry date fig grape]
}

// --- example_keys_test.go (custom sort.Interface using a key function) ---

// A couple of type definitions to make the units clear.
type earthMass float64
type au float64

// A Planet defines the properties of a solar system object.
type Planet struct {
	name     string
	mass     earthMass
	distance au
}

// By is the type of a "less" function that defines the ordering of its
// Planet arguments.
type By func(p1, p2 *Planet) bool

// Sort is a method on the function type, By, that sorts the argument slice
// according to the function.
func (by By) Sort(planets []Planet) {
	ps := &planetSorter{
		planets: planets,
		by:      by,
	}
	sort.Sort(ps)
}

// planetSorter joins a By function and a slice of Planets to be sorted.
type planetSorter struct {
	planets []Planet
	by      func(p1, p2 *Planet) bool
}

func (s *planetSorter) Len() int      { return len(s.planets) }
func (s *planetSorter) Swap(i, j int) { s.planets[i], s.planets[j] = s.planets[j], s.planets[i] }
func (s *planetSorter) Less(i, j int) bool {
	return s.by(&s.planets[i], &s.planets[j])
}

var planets = []Planet{
	{"Mercury", 0.055, 0.4},
	{"Venus", 0.815, 0.7},
	{"Earth", 1.0, 1.0},
	{"Mars", 0.107, 1.5},
}

func Example_sortKeys() {
	name := func(p1, p2 *Planet) bool { return p1.name < p2.name }
	mass := func(p1, p2 *Planet) bool { return p1.mass < p2.mass }
	distance := func(p1, p2 *Planet) bool { return p1.distance < p2.distance }
	decreasingDistance := func(p1, p2 *Planet) bool { return distance(p2, p1) }

	By(name).Sort(planets)
	fmt.Println("By name:", planets)

	By(mass).Sort(planets)
	fmt.Println("By mass:", planets)

	By(distance).Sort(planets)
	fmt.Println("By distance:", planets)

	By(decreasingDistance).Sort(planets)
	fmt.Println("By decreasing distance:", planets)

	// Output: By name: [{Earth 1 1} {Mars 0.107 1.5} {Mercury 0.055 0.4} {Venus 0.815 0.7}]
	// By mass: [{Mercury 0.055 0.4} {Mars 0.107 1.5} {Venus 0.815 0.7} {Earth 1 1}]
	// By distance: [{Mercury 0.055 0.4} {Venus 0.815 0.7} {Earth 1 1} {Mars 0.107 1.5}]
	// By decreasing distance: [{Mars 0.107 1.5} {Earth 1 1} {Venus 0.815 0.7} {Mercury 0.055 0.4}]
}

// --- example_multi_test.go ---

type Change struct {
	user     string
	language string
	lines    int
}

type lessFunc func(p1, p2 *Change) bool

type multiSorter struct {
	changes []Change
	less    []lessFunc
}

func (ms *multiSorter) Sort(changes []Change) {
	ms.changes = changes
	sort.Sort(ms)
}

func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{less: less}
}

func (ms *multiSorter) Len() int      { return len(ms.changes) }
func (ms *multiSorter) Swap(i, j int) { ms.changes[i], ms.changes[j] = ms.changes[j], ms.changes[i] }
func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.changes[i], &ms.changes[j]
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return ms.less[k](p, q)
}

var changes = []Change{
	{"gri", "Go", 100},
	{"ken", "C", 150},
	{"glenda", "Go", 200},
	{"rsc", "Go", 200},
	{"r", "Go", 100},
	{"ken", "Go", 200},
	{"dmr", "C", 100},
	{"r", "C", 150},
	{"gri", "Smalltalk", 80},
}

func Example_sortMultiKeys() {
	user := func(c1, c2 *Change) bool { return c1.user < c2.user }
	language := func(c1, c2 *Change) bool { return c1.language < c2.language }
	increasingLines := func(c1, c2 *Change) bool { return c1.lines < c2.lines }
	decreasingLines := func(c1, c2 *Change) bool { return c1.lines > c2.lines }

	OrderedBy(user).Sort(changes)
	fmt.Println("By user:", changes)

	OrderedBy(user, increasingLines).Sort(changes)
	fmt.Println("By user,<lines:", changes)

	OrderedBy(user, decreasingLines).Sort(changes)
	fmt.Println("By user,>lines:", changes)

	OrderedBy(language, increasingLines).Sort(changes)
	fmt.Println("By language,<lines:", changes)

	OrderedBy(language, increasingLines, user).Sort(changes)
	fmt.Println("By language,<lines,user:", changes)

	// Output:
	// By user: [{dmr C 100} {glenda Go 200} {gri Go 100} {gri Smalltalk 80} {ken C 150} {ken Go 200} {r Go 100} {r C 150} {rsc Go 200}]
	// By user,<lines: [{dmr C 100} {glenda Go 200} {gri Smalltalk 80} {gri Go 100} {ken C 150} {ken Go 200} {r Go 100} {r C 150} {rsc Go 200}]
	// By user,>lines: [{dmr C 100} {glenda Go 200} {gri Go 100} {gri Smalltalk 80} {ken Go 200} {ken C 150} {r C 150} {r Go 100} {rsc Go 200}]
	// By language,<lines: [{dmr C 100} {ken C 150} {r C 150} {gri Go 100} {r Go 100} {glenda Go 200} {ken Go 200} {rsc Go 200} {gri Smalltalk 80}]
	// By language,<lines,user: [{dmr C 100} {ken C 150} {r C 150} {gri Go 100} {r Go 100} {glenda Go 200} {ken Go 200} {rsc Go 200} {gri Smalltalk 80}]
}

// --- example_wrapper_test.go ---

type Grams int

func (g Grams) String() string { return fmt.Sprintf("%dg", int(g)) }

type Organ struct {
	Name   string
	Weight Grams
}

type Organs []*Organ

func (s Organs) Len() int      { return len(s) }
func (s Organs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByName struct{ Organs }

func (s ByName) Less(i, j int) bool { return s.Organs[i].Name < s.Organs[j].Name }

type ByWeight struct{ Organs }

func (s ByWeight) Less(i, j int) bool { return s.Organs[i].Weight < s.Organs[j].Weight }

func Example_sortWrapper() {
	s := []*Organ{
		{"brain", 1340},
		{"heart", 290},
		{"liver", 1494},
		{"pancreas", 131},
		{"prostate", 62},
		{"spleen", 162},
	}

	sort.Sort(ByWeight{s})
	fmt.Println("Organs by weight:")
	printOrgans(s)

	sort.Sort(ByName{s})
	fmt.Println("Organs by name:")
	printOrgans(s)

	// Output:
	// Organs by weight:
	// prostate (62g)
	// pancreas (131g)
	// spleen   (162g)
	// heart    (290g)
	// brain    (1340g)
	// liver    (1494g)
	// Organs by name:
	// brain    (1340g)
	// heart    (290g)
	// liver    (1494g)
	// pancreas (131g)
	// prostate (62g)
	// spleen   (162g)
}

func printOrgans(s []*Organ) {
	for _, o := range s {
		fmt.Printf("%-8s (%v)\n", o.Name, o.Weight)
	}
}
