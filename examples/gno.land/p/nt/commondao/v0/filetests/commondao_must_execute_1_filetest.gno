// PKGPATH: gno.land/r/test/commondao_must_execute
package commondao_must_execute

import (
	"errors"
	"testing"
	"time"

	"gno.land/p/nt/commondao/v0"
)

type mockPropDef struct{}

func (mockPropDef) Title() string                               { return "" }
func (mockPropDef) Body() string                                { return "" }
func (mockPropDef) VotingPeriod() time.Duration                 { return 0 }
func (mockPropDef) Tally(commondao.VotingContext) (bool, error) { return true, nil }

func (mockPropDef) Executor() commondao.ExecFunc {
	return func(realm) error {
		return errors.New("boom!")
	}
}

func main(cur realm) {
	// Make sure proposal is executed in a realm different than the one where definition has been defined
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/testing/dao"))

	commondao.MustExecute(mockPropDef{}, cur)
}

// Error:
// boom!
