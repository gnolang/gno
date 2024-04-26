package emitter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	t.Parallel()

	svr := NewServer(log.NewTestingLogger(t))

	s := httptest.NewServer(http.HandlerFunc(svr.ServeHTTP))
	defer s.Close()

	u := "ws" + strings.TrimPrefix(s.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("client Dial failed: %v", err)
	}
	defer c.Close()

	sendEvt := events.Custom("TEST")
	assert.Len(t, svr.clients, 1)
	svr.Emit(sendEvt) // simulate reload

	var recvEvt eventJSON
	err = c.ReadJSON(&recvEvt)
	require.NoError(t, err)
	assert.Equal(t, sendEvt.Type(), recvEvt.Type)
}
