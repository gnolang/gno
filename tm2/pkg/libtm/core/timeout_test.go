package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeout_CalculateTimeout(t *testing.T) {
	t.Parallel()

	var (
		initial = 10 * time.Second
		delta   = 200 * time.Millisecond

		tm = Timeout{
			Initial: initial,
			Delta:   delta,
		}
	)

	for round := uint64(0); round < 100; round++ {
		assert.Equal(
			t,
			initial+time.Duration(round)*delta,
			tm.CalculateTimeout(round),
		)
	}
}

func TestTimeout_ScheduleTimeoutPropose(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name string
		step step
	}{
		{
			"OnTimeoutPropose",
			propose,
		},
		{
			"OnTimeoutPrevote",
			prevote,
		},
		{
			"OnTimeoutPrecommit",
			precommit,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			expiredCh := make(chan struct{}, 1)

			tm := NewTendermint(
				nil,
				nil,
				nil,
				nil,
			)

			// Set the timeout data for the step
			tm.timeouts[testCase.step] = Timeout{
				Initial: 50 * time.Millisecond,
				Delta:   50 * time.Millisecond,
			}

			// Schedule the timeout
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			timeoutPropose := tm.timeouts[testCase.step].CalculateTimeout(0)

			tm.scheduleTimeout(ctx, timeoutPropose, expiredCh)

			// Wait for the timer to trigger
			select {
			case <-time.After(5 * time.Second):
				t.Fatal("timer not triggered")
			case <-expiredCh:
			}

			tm.wg.Wait()
		})
	}
}
