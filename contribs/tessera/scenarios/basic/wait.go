package basic

import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gno/contribs/tessera/pkg/cluster"
	"github.com/gnolang/gno/contribs/tessera/pkg/scenario"
)

func init() {
	scenario.RegisterScenario(
		"wait_for_height",
		&WaitForHeight{
			Height:  1,
			Timeout: 30 * time.Second,
		},
	)
}

type WaitForHeight struct {
	Height  int64         `yaml:"height"`
	Timeout time.Duration `yaml:"timeout"`
}

func (w *WaitForHeight) Description() string {
	return fmt.Sprintf(
		"Wait for height %d with timeout %.2fs",
		w.Height,
		w.Timeout.Seconds(),
	)
}

func (w *WaitForHeight) Execute(execCtx context.Context, _ *cluster.Cluster) error {
	_, cancelFn := context.WithTimeout(execCtx, w.Timeout)
	defer cancelFn()

	// TODO implement

	return nil
}

func (w *WaitForHeight) Verify(ctx context.Context, _ *cluster.Cluster) error {
	// No verification needed
	return nil
}

// retryUntilTimeout runs the callback until the timeout is exceeded, or
// the callback returns a flag indicating completion
func retryUntilTimeout(ctx context.Context, cb func() bool) error {
	ch := make(chan error, 1)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				ch <- ctx.Err()

				return
			default:
				retry := cb()
				if !retry {
					ch <- nil
					return
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	return <-ch
}
