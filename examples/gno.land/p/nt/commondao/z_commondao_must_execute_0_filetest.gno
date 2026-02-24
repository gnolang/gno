// PKGPATH: gno.land/r/test
package test

import (
	"testing"
	"time"

	"gno.land/p/nt/commondao"
)

type testPropDef struct{}

func (testPropDef) Title() string                               { return "" }
func (testPropDef) Body() string                                { return "" }
func (testPropDef) VotingPeriod() time.Duration                 { return 0 }
func (testPropDef) Executor() commondao.ExecFunc                { return nil }
func (testPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func main() {
	// Make sure proposal is executed in a realm different than the one where definitions has been defined
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/testing/dao"))

	commondao.MustExecute(testPropDef{})
	println("ok")
}

// Output:
// ok
