// Package launderrdeep stores slices below the top level of its values, for
// the filetests that convert them from another realm.
package launderrdeep

type Box struct{ Items []string }

type Outer struct{ Mid struct{ Items []string } }

var (
	grid  = [][]string{{"a", "b"}, {"c", "d"}}
	box   = &Box{Items: []string{"e", "f"}}
	table = map[string][][]string{"k": {{"g", "h"}}}
	outer = Outer{Mid: struct{ Items []string }{Items: []string{"i", "j"}}}
)

func Grid() [][]string             { return grid }
func GetBox() *Box                 { return box }
func Table() map[string][][]string { return table }
func GetOuter() Outer              { return outer }
func GetOuterPtr() *Outer          { return &outer }
