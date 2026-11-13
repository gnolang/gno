package state

import (
	"sync"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// validatorsCacheInterval is how many heights apart ValidatorsCache keeps a
// replayed set, which bounds the rounds a load replays once the segment is kept.
const validatorsCacheInterval = 1000

// ValidatorsCache returns what LoadValidators returns, keeping a replayed set
// every validatorsCacheInterval heights so a load resumes from the nearest one.
// The zero value is ready to use.
type ValidatorsCache struct {
	mtx sync.Mutex
	// db is the DB every kept set was replayed from.
	db dbm.DB
	// kept maps {stored height, rounds} to the set stored at that height
	// after that many rounds.
	kept map[[2]int64]*types.ValidatorSet
}

// LoadValidators returns the validator set at height, equal to what the
// package-level LoadValidators returns.
func (c *ValidatorsCache) LoadValidators(db dbm.DB, height int64) (*types.ValidatorSet, error) {
	vals, _, err := c.loadValidators(db, height)
	return vals, err
}

// loadValidators also returns the number of proposer rounds it replayed.
func (c *ValidatorsCache) loadValidators(db dbm.DB, height int64) (*types.ValidatorSet, int64, error) {
	vals, stored, err := loadStoredValidators(db, height)
	if err != nil || stored == height {
		return vals, 0, err
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.kept == nil || c.db != db {
		c.db, c.kept = db, make(map[[2]int64]*types.ValidatorSet)
	}

	// Rounds are counted from the stored set. Resume from the last kept set
	// below the target, so at least one round runs and sets the proposer.
	target := height - stored
	done := int64(0)
	for r := (target - 1) / validatorsCacheInterval * validatorsCacheInterval; r > 0; r -= validatorsCacheInterval {
		if kept, ok := c.kept[[2]int64{stored, r}]; ok {
			vals, done = kept.Copy(), r
			break
		}
	}
	replayed := target - done
	if done == 0 {
		// The first round rescales and centres the priorities, as
		// LoadValidators does before replaying.
		vals.IncrementProposerPriority(1)
		done = 1
	}
	for next := (done/validatorsCacheInterval + 1) * validatorsCacheInterval; next < target; next += validatorsCacheInterval {
		vals.AdvanceProposerPriority(int(next - done))
		done = next
		c.kept[[2]int64{stored, done}] = vals.Copy()
	}
	if done < target {
		vals.AdvanceProposerPriority(int(target - done))
	}
	return vals, replayed, nil
}
