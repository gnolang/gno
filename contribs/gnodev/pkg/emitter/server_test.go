package emitter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_New(t *testing.T) {
	t.Parallel()

	svr := NewServer(log.NewTestingLogger(t))
	assert.Len(t, svr.clients, 0)
}

func TestServer_ServeHTTP(t *testing.T) {
	svr := NewServer(log.NewTestingLogger(t))

	s := httptest.NewServer(http.HandlerFunc(svr.ServeHTTP))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err, "client Dial failed")

	defer c.Close()

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.Len(c, svr.conns(), 1)
	}, time.Second, time.Millisecond*100)

	sendEvt := events.Custom("TEST")
	svr.Emit(sendEvt) // simulate reload

	var recvEvt eventJSON
	err = c.ReadJSON(&recvEvt)
	require.NoError(t, err)
	assert.Equal(t, sendEvt.Type(), recvEvt.Type)
}
