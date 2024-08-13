package server

import (
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gnostats/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubs_Subscribe(t *testing.T) {
	t.Parallel()

	s := make(subs)

	id, ch := s.subscribe()
	require.NotNil(t, id)
	require.NotNil(t, ch)

	assert.Len(t, s, 1)
}

func TestSubs_Unsubscribe(t *testing.T) {
	t.Parallel()

	s := make(subs)

	id, _ := s.subscribe()
	require.NotNil(t, id)

	require.Len(t, s, 1)

	s.unsubscribe(id)
	assert.Len(t, s, 0)
}

func TestSubs_Notify(t *testing.T) {
	t.Parallel()

	var (
		s             = make(subs)
		receivedPoint *proto.DataPoint

		dataPoint = generateDataPoints(t, 1)[0]
	)

	id, ch := s.subscribe()
	require.NotNil(t, id)
	require.NotNil(t, ch)

	defer s.unsubscribe(id)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		select {
		case <-time.After(5 * time.Second):
		case receivedPoint = <-ch:
		}
	}()

	s.notify(dataPoint)

	wg.Wait()

	assert.Equal(t, dataPoint.StaticInfo.Address, receivedPoint.StaticInfo.Address)
	assert.Equal(t, dataPoint.StaticInfo.GnoVersion, receivedPoint.StaticInfo.GnoVersion)
	assert.Equal(t, dataPoint.StaticInfo.OsVersion, receivedPoint.StaticInfo.OsVersion)
}
