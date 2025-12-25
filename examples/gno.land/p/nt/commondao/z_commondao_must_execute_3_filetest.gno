// PKGPATH: gno.land/r/test
package test

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/nt/commondao"
)

type testCrossPropDef struct{}

func (testCrossPropDef) Title() string                               { return "" }
func (testCrossPropDef) Body() string                                { return "" }
func (testCrossPropDef) VotingPeriod() time.Duration                 { return 0 }
func (testCrossPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }
func (testCrossPropDef) Execute(realm) error                         { return errors.New("boom!") }

func main() {
	// Make sure proposal is executed in a realm different than the one where definition has been defined
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/testing/dao"))

	commondao.MustExecute(testCrossPropDef{})
}

// Error:
// boom!
