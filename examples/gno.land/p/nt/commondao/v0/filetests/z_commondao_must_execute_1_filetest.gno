// PKGPATH: gno.land/r/test
package test

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/nt/commondao/v0"
)

type testPropDef struct{}

func (testPropDef) Title() string                               { return "" }
func (testPropDef) Body() string                                { return "" }
func (testPropDef) VotingPeriod() time.Duration                 { return 0 }
func (testPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func (testPropDef) Executor() commondao.ExecFunc {
	return func(realm) error {
		return errors.New("boom!")
	}
}

func main() {
	// Make sure proposal is executed in a realm different than the one where definition has been defined
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/testing/dao"))

	commondao.MustExecute(testPropDef{})
}

// Error:
// boom!
