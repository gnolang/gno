package core

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTendermint_Options(t *testing.T) {
	t.Parallel()

	t.Run("Withlogger", func(t *testing.T) {
		t.Parallel()

		l := slog.New(slog.NewTextHandler(io.Discard, nil))

		tm := NewTendermint(
			nil,
			nil,
			nil,
			nil,
			WithLogger(l),
		)

		assert.Equal(t, tm.logger, l)
	})

	t.Run("WithProposeTimeout", func(t *testing.T) {
		t.Parallel()

		timeout := Timeout{
			Initial: 500 * time.Millisecond,
			Delta:   0,
		}

		tm := NewTendermint(
			nil,
			nil,
			nil,
			nil,
			WithProposeTimeout(timeout),
		)

		assert.Equal(t, tm.timeouts[propose], timeout)
	})

	t.Run("WithPrevoteTimeout", func(t *testing.T) {
		t.Parallel()

		timeout := Timeout{
			Initial: 500 * time.Millisecond,
			Delta:   0,
		}

		tm := NewTendermint(
			nil,
			nil,
			nil,
			nil,
			WithPrevoteTimeout(timeout),
		)

		assert.Equal(t, tm.timeouts[prevote], timeout)
	})

	t.Run("WithPrecommitTimeout", func(t *testing.T) {
		t.Parallel()

		timeout := Timeout{
			Initial: 500 * time.Millisecond,
			Delta:   0,
		}

		tm := NewTendermint(
			nil,
			nil,
			nil,
			nil,
			WithPrecommitTimeout(timeout),
		)

		assert.Equal(t, tm.timeouts[precommit], timeout)
	})
}
