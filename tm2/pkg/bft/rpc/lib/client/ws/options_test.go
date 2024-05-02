package ws

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_WithLogger(t *testing.T) {
	t.Parallel()

	var (
		upgrader = websocket.Upgrader{}

		handler = func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrader.Upgrade(w, r, nil)

			require.NoError(t, err)
			require.NoError(t, c.Close())
		}
	)

	s := createTestServer(t, http.HandlerFunc(handler))
	url := "ws" + strings.TrimPrefix(s.URL, "http")

	// Create the client
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c, err := NewClient(url, WithLogger(logger))
	require.NoError(t, err)

	assert.Equal(t, logger, c.logger)
}
