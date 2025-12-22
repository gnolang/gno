// PKGPATH: gno.land/r/test
package test

import (
	"time"

	"gno.land/p/nt/commondao"
)

type testPropDef struct{}

func (testPropDef) Title() string                               { return "" }
func (testPropDef) Body() string                                { return "" }
func (testPropDef) VotingPeriod() time.Duration                 { return 0 }
func (testPropDef) Execute(cur realm) error                     { return nil }
func (testPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func main() {
	commondao.MustExecute(testPropDef{})
	println("ok")
}

// Output:
// ok
