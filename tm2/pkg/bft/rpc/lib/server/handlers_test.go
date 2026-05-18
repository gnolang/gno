package rpcserver_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rs "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// -----------
// HTTP REST API
// TODO

// -----------
// JSON-RPC over HTTP

func testMux() *http.ServeMux {
	funcMap := map[string]*rs.RPCFunc{
		"c": rs.NewRPCFunc(func(ctx *types.Context, s string, i int) (string, error) { return "foo", nil }, "s,i"),
	}
	mux := http.NewServeMux()

	rs.RegisterRPCFuncs(mux, funcMap, log.NewNoopLogger())

	return mux
}

func statusOK(code int) bool { return code >= 200 && code <= 299 }

// Ensure that nefarious/unintended inputs to `params`
// do not crash our RPC handlers.
// See Issue https://github.com/tendermint/tendermint/issues/708.
func TestRPCParams(t *testing.T) {
	t.Parallel()

	mux := testMux()
	tests := []struct {
		payload    string
		wantErr    string
		expectedId any
	}{
		// bad
		{`{"jsonrpc": "2.0", "id": "0"}`, "Method not found", types.JSONRPCStringID("0")},
		{`{"jsonrpc": "2.0", "method": "y", "id": "0"}`, "Method not found", types.JSONRPCStringID("0")},
		{`{"method": "c", "id": "0", "params": a}`, "invalid character", types.JSONRPCStringID("")}, // id not captured in JSON parsing failures
		{`{"method": "c", "id": "0", "params": ["a"]}`, "got 1", types.JSONRPCStringID("0")},
		{`{"method": "c", "id": "0", "params": ["a", "b"]}`, "invalid character", types.JSONRPCStringID("0")},
		{`{"method": "c", "id": "0", "params": [1, 1]}`, "of type string", types.JSONRPCStringID("0")},

		// good
		{`{"jsonrpc": "2.0", "method": "c", "id": "0", "params": null}`, "", types.JSONRPCStringID("0")},
		{`{"method": "c", "id": "0", "params": {}}`, "", types.JSONRPCStringID("0")},
		{`{"method": "c", "id": "0", "params": ["a", "10"]}`, "", types.JSONRPCStringID("0")},
	}

	for i, tt := range tests {
		req, _ := http.NewRequest("POST", "http://localhost/", strings.NewReader(tt.payload))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		res := rec.Result()
		// Always expecting back a JSONRPCResponse
		assert.True(t, statusOK(res.StatusCode), "#%d: should always return 2XX", i)
		blob, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("#%d: err reading body: %v", i, err)
			continue
		}

		recv := new(types.RPCResponse)
		assert.Nil(t, json.Unmarshal(blob, recv), "#%d: expecting successful parsing of an RPCResponse:\nblob: %s", i, blob)
		assert.NotEqual(t, recv, new(types.RPCResponse), "#%d: not expecting a blank RPCResponse", i)
		assert.Equal(t, tt.expectedId, recv.ID, "#%d: expected ID not matched in RPCResponse", i)
		if tt.wantErr == "" {
			assert.Nil(t, recv.Error, "#%d: not expecting an error", i)
		} else {
			assert.True(t, recv.Error.Code < 0, "#%d: not expecting a positive JSONRPC code", i)
			// The wanted error is either in the message or the data
			assert.Contains(t, recv.Error.Message+recv.Error.Data, tt.wantErr, "#%d: expected substring", i)
		}
	}
}

// streamableTestResult is an RPC result type that implements StreamableResult
// to exercise the streaming code path end-to-end through the registered
// HTTP handler.
type streamableTestResult struct {
	Greeting string `json:"greeting"`
}

func (r *streamableTestResult) StreamJSON(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, `{"greeting":"`+r.Greeting+`","streamed":true}`)
	return err
}

func streamingTestMux() *http.ServeMux {
	funcMap := map[string]*rs.RPCFunc{
		"stream": rs.NewRPCFunc(func(ctx *types.Context) (*streamableTestResult, error) {
			return &streamableTestResult{Greeting: "hello"}, nil
		}, ""),
	}
	mux := http.NewServeMux()
	rs.RegisterRPCFuncs(mux, funcMap, log.NewNoopLogger())
	return mux
}

// TestStreamableResult_HTTPGet verifies that GET /<method> on a method that
// returns a StreamableResult writes the streamed body inside the JSON-RPC
// envelope, bypassing the standard amino+MarshalIndent path.
func TestStreamableResult_HTTPGet(t *testing.T) {
	mux := streamingTestMux()

	req, _ := http.NewRequest("GET", "http://localhost/stream", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	res := rec.Result()
	require.Equal(t, 200, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	// "streamed":true is a marker only StreamJSON emits; the standard
	// reflection-based marshal path would not produce it because
	// streamableTestResult.Greeting is the only exported JSON field.
	require.Contains(t, string(body), `"streamed":true`,
		"streamed body must reach the wire (got %s)", body)
	require.Contains(t, string(body), `"greeting":"hello"`)
	require.Contains(t, string(body), `"jsonrpc":"2.0"`)
}

// TestStreamableResult_JSONRPCPost verifies that a single JSON-RPC POST whose
// method returns a StreamableResult also writes the streamed body inside the
// envelope. This is the path used by the Go RPC client SDK.
func TestStreamableResult_JSONRPCPost(t *testing.T) {
	mux := streamingTestMux()

	payload := `{"jsonrpc":"2.0","method":"stream","id":"42","params":{}}`
	req, _ := http.NewRequest("POST", "http://localhost/", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	res := rec.Result()
	require.Equal(t, 200, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Contains(t, string(body), `"streamed":true`,
		"streamed body must reach the wire on the JSON-RPC POST path (got %s)", body)
	require.Contains(t, string(body), `"id":"42"`)
}

// TestStreamableResult_BatchPostReturnsError verifies that a batched JSON-RPC
// POST whose method returns a StreamableResult does NOT silently drop the
// response slot. Because streaming inside a JSON-array batch would interleave
// bodies with no way for the client to demarcate them, the handler must emit
// an explicit error for the streamable slot rather than skipping it.
//
// Pre-fix: the batch loop did `resp, _ := processRequest(...)` and silently
// discarded the streamable, leaving the client with a batch response that
// had fewer entries than the request — no error, no log line.
func TestStreamableResult_BatchPostReturnsError(t *testing.T) {
	mux := streamingTestMux()

	// Batch of two requests: a streamable method and a non-existent method
	// (to confirm the non-streamable error path still works alongside).
	payload := `[
		{"jsonrpc":"2.0","method":"stream","id":"1","params":{}},
		{"jsonrpc":"2.0","method":"missing","id":"2","params":{}}
	]`
	req, _ := http.NewRequest("POST", "http://localhost/", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	res := rec.Result()
	require.Equal(t, 200, res.StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	var resps []map[string]any
	require.NoError(t, json.Unmarshal(body, &resps))
	require.Len(t, resps, 2,
		"batch response must contain one entry per request, got %d (%s)", len(resps), body)

	// First slot (the streamable method) must surface an explicit JSON-RPC
	// error — not be silently absent or contain a partial streamed body.
	first := resps[0]
	assert.Equal(t, "1", first["id"], "first slot must keep the request id")
	require.Contains(t, first, "error",
		"first slot must carry a JSON-RPC error explaining streaming is unsupported in batch (got %v)", first)
}

func TestJSONRPCID(t *testing.T) {
	t.Parallel()

	mux := testMux()
	tests := []struct {
		payload    string
		wantErr    bool
		expectedId any
	}{
		// good id
		{`{"jsonrpc": "2.0", "method": "c", "id": "0", "params": ["a", "10"]}`, false, types.JSONRPCStringID("0")},
		{`{"jsonrpc": "2.0", "method": "c", "id": "abc", "params": ["a", "10"]}`, false, types.JSONRPCStringID("abc")},
		{`{"jsonrpc": "2.0", "method": "c", "id": 0, "params": ["a", "10"]}`, false, types.JSONRPCIntID(0)},
		{`{"jsonrpc": "2.0", "method": "c", "id": 1, "params": ["a", "10"]}`, false, types.JSONRPCIntID(1)},
		{`{"jsonrpc": "2.0", "method": "c", "id": 1.3, "params": ["a", "10"]}`, false, types.JSONRPCIntID(1)},
		{`{"jsonrpc": "2.0", "method": "c", "id": -1, "params": ["a", "10"]}`, false, types.JSONRPCIntID(-1)},

		// bad id
		{`{"jsonrpc": "2.0", "method": "c", "id": null, "params": ["a", "10"]}`, true, nil},
		{`{"jsonrpc": "2.0", "method": "c", "id": {}, "params": ["a", "10"]}`, true, nil},
		{`{"jsonrpc": "2.0", "method": "c", "id": [], "params": ["a", "10"]}`, true, nil},
	}

	for i, tt := range tests {
		req, _ := http.NewRequest("POST", "http://localhost/", strings.NewReader(tt.payload))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		res := rec.Result()
		// Always expecting back a JSONRPCResponse
		assert.True(t, statusOK(res.StatusCode), "#%d: should always return 2XX", i)
		blob, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("#%d: err reading body: %v", i, err)
			continue
		}

		recv := new(types.RPCResponse)
		err = json.Unmarshal(blob, recv)
		assert.Nil(t, err, "#%d: expecting successful parsing of an RPCResponse:\nblob: %s", i, blob)
		if !tt.wantErr {
			assert.NotEqual(t, recv, new(types.RPCResponse), "#%d: not expecting a blank RPCResponse", i)
			assert.Equal(t, tt.expectedId, recv.ID, "#%d: expected ID not matched in RPCResponse", i)
			assert.Nil(t, recv.Error, "#%d: not expecting an error", i)
		} else {
			assert.True(t, recv.Error.Code < 0, "#%d: not expecting a positive JSONRPC code", i)
		}
	}
}

func TestRPCNotification(t *testing.T) {
	t.Parallel()

	mux := testMux()
	body := strings.NewReader(`{"jsonrpc": "2.0", "id": ""}`)
	req, _ := http.NewRequest("POST", "http://localhost/", body)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	res := rec.Result()

	// Always expecting back a JSONRPCResponse
	require.True(t, statusOK(res.StatusCode), "should always return 2XX")
	blob, err := io.ReadAll(res.Body)
	require.Nil(t, err, "reading from the body should not give back an error")
	require.Equal(t, len(blob), 0, "a notification SHOULD NOT be responded to by the server")
}

func TestRPCNotificationInBatch(t *testing.T) {
	t.Parallel()

	mux := testMux()
	tests := []struct {
		payload     string
		expectCount int
	}{
		{
			`[
				{"jsonrpc": "2.0","id": ""},
				{"jsonrpc": "2.0","method":"c","id":"abc","params":["a","10"]}
			 ]`,
			1,
		},
		{
			`[
				{"jsonrpc": "2.0","method":"c","id":"abc","params":["a","10"]}
			 ]`,
			1,
		},
		{
			`[
				{"jsonrpc": "2.0","id": ""},
				{"jsonrpc": "2.0","method":"c","id":"abc","params":["a","10"]},
				{"jsonrpc": "2.0","id": ""},
				{"jsonrpc": "2.0","method":"c","id":"abc","params":["a","10"]}
			 ]`,
			2,
		},
	}
	for i, tt := range tests {
		req, _ := http.NewRequest("POST", "http://localhost/", strings.NewReader(tt.payload))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		res := rec.Result()
		// Always expecting back a JSONRPCResponse
		assert.True(t, statusOK(res.StatusCode), "#%d: should always return 2XX", i)
		blob, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("#%d: err reading body: %v", i, err)
			continue
		}

		var responses types.RPCResponses
		// try to unmarshal an array first
		err = json.Unmarshal(blob, &responses)
		if err != nil {
			t.Errorf("#%d: expected an array, couldn't unmarshal it\nblob: %s", i, blob)
			continue
		}
		if tt.expectCount != len(responses) {
			t.Errorf("#%d: expected %d response(s), but got %d\nblob: %s", i, tt.expectCount, len(responses), blob)
			continue
		}
		for _, response := range responses {
			assert.NotEqual(t, response, new(types.RPCResponse), "#%d: not expecting a blank RPCResponse", i)
		}
	}
}

func TestRPCBatchPartialUnmarshal(t *testing.T) {
	t.Parallel()

	mux := testMux()
	tests := []struct {
		name        string
		payload     string
		expectCount int
	}{
		{
			// One valid request + one with null id (fails UnmarshalJSON).
			// Valid request should succeed; malformed one should get an error response.
			name: "null_id_does_not_drop_batch",
			payload: `[
				{"jsonrpc":"2.0","method":"c","id":"1","params":["a","10"]},
				{"jsonrpc":"2.0","method":"c","id":null,"params":["a","10"]}
			]`,
			expectCount: 2,
		},
		{
			// One valid request + one with missing id field.
			name: "missing_id_does_not_drop_batch",
			payload: `[
				{"jsonrpc":"2.0","method":"c","id":"1","params":["a","10"]},
				{"jsonrpc":"2.0","method":"c","params":["a","10"]}
			]`,
			expectCount: 2,
		},
		{
			// One valid request + one with object id (fails parseID).
			name: "object_id_does_not_drop_batch",
			payload: `[
				{"jsonrpc":"2.0","method":"c","id":"1","params":["a","10"]},
				{"jsonrpc":"2.0","method":"c","id":{},"params":["a","10"]}
			]`,
			expectCount: 2,
		},
		{
			// Three valid requests + one malformed in the middle.
			name: "malformed_element_preserves_others",
			payload: `[
				{"jsonrpc":"2.0","method":"c","id":"1","params":["a","10"]},
				{"jsonrpc":"2.0","method":"c","id":null,"params":["a","10"]},
				{"jsonrpc":"2.0","method":"c","id":"2","params":["a","10"]},
				{"jsonrpc":"2.0","method":"c","id":"3","params":["a","10"]}
			]`,
			expectCount: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest("POST", "http://localhost/", strings.NewReader(tt.payload))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			res := rec.Result()

			assert.True(t, statusOK(res.StatusCode), "should always return 2XX")
			blob, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, blob, "response body should not be empty — batch must not be silently dropped")

			var responses types.RPCResponses
			require.NoError(t, json.Unmarshal(blob, &responses), "response should be a JSON array, got: %s", blob)
			require.Len(t, responses, tt.expectCount, "unexpected number of responses\nblob: %s", blob)

			// Verify we get at least one success and at least one error
			var hasSuccess, hasError bool
			for _, resp := range responses {
				if resp.Error != nil {
					hasError = true
				} else {
					hasSuccess = true
				}
			}
			assert.True(t, hasSuccess, "expected at least one successful response")
			assert.True(t, hasError, "expected at least one error response for malformed request")
		})
	}
}

func TestUnknownRPCPath(t *testing.T) {
	t.Parallel()

	mux := testMux()
	req, _ := http.NewRequest("GET", "http://localhost/unknownrpcpath", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	res := rec.Result()

	// Always expecting back a 404 error
	require.Equal(t, http.StatusNotFound, res.StatusCode, "should always return 404")
}

// -----------
// JSON-RPC over WEBSOCKETS

func TestWebsocketManagerHandler(t *testing.T) {
	t.Parallel()

	s := newWSServer()
	defer s.Close()

	// check upgrader works
	d := websocket.Dialer{}
	c, dialResp, err := d.Dial("ws://"+s.Listener.Addr().String()+"/websocket", nil)
	require.NoError(t, err)

	if got, want := dialResp.StatusCode, http.StatusSwitchingProtocols; got != want {
		t.Errorf("dialResp.StatusCode = %q, want %q", got, want)
	}

	// check basic functionality works
	req, err := types.MapToRequest(types.JSONRPCStringID("TestWebsocketManager"), "c", map[string]any{"s": "a", "i": 10})
	require.NoError(t, err)
	err = c.WriteJSON(req)
	require.NoError(t, err)

	var resp types.RPCResponse
	err = c.ReadJSON(&resp)
	require.NoError(t, err)
	require.Nil(t, resp.Error)
}

func newWSServer() *httptest.Server {
	funcMap := map[string]*rs.RPCFunc{
		"c": rs.NewWSRPCFunc(func(ctx *types.Context, s string, i int) (string, error) { return "foo", nil }, "s,i"),
	}
	wm := rs.NewWebsocketManager(funcMap)
	wm.SetLogger(log.NewNoopLogger())

	mux := http.NewServeMux()
	mux.HandleFunc("/websocket", wm.WebsocketHandler)

	return httptest.NewServer(mux)
}

func newWSStreamingServer() *httptest.Server {
	funcMap := map[string]*rs.RPCFunc{
		"stream": rs.NewRPCFunc(func(ctx *types.Context) (*streamableTestResult, error) {
			return &streamableTestResult{Greeting: "hello"}, nil
		}, ""),
		"echo": rs.NewRPCFunc(func(ctx *types.Context) (string, error) {
			return "pong", nil
		}, ""),
	}
	wm := rs.NewWebsocketManager(funcMap)
	wm.SetLogger(log.NewNoopLogger())

	mux := http.NewServeMux()
	mux.HandleFunc("/websocket", wm.WebsocketHandler)

	return httptest.NewServer(mux)
}

// TestStreamableResult_WebSocketStreams verifies that a single WebSocket
// JSON-RPC request whose method returns a StreamableResult streams the body
// via NextWriter rather than buffering it, and that the client receives the
// streamed content inside a valid JSON-RPC envelope.
func TestStreamableResult_WebSocketStreams(t *testing.T) {
	t.Parallel()

	s := newWSStreamingServer()
	defer s.Close()

	d := websocket.Dialer{}
	c, _, err := d.Dial("ws://"+s.Listener.Addr().String()+"/websocket", nil)
	require.NoError(t, err)
	defer c.Close()

	req, err := types.MapToRequest(types.JSONRPCStringID("ws-stream-test"), "stream", map[string]any{})
	require.NoError(t, err)
	err = c.WriteJSON(req)
	require.NoError(t, err)

	var resp types.RPCResponse
	err = c.ReadJSON(&resp)
	require.NoError(t, err)

	require.Nil(t, resp.Error,
		"WebSocket single-request streaming must succeed, got error: %v", resp.Error)
	require.Equal(t, types.JSONRPCStringID("ws-stream-test"), resp.ID)

	// "streamed":true is only emitted by StreamJSON, not by the standard
	// json.MarshalIndent path, so its presence proves the streaming path ran.
	raw, err := json.Marshal(resp.Result)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"streamed":true`,
		"result body must come from StreamJSON (got %s)", raw)
	require.Contains(t, string(raw), `"greeting":"hello"`)
}

// TestStreamableResult_WebSocketBatchReturnsError verifies that a batch
// WebSocket request containing a streamable method returns an explicit
// JSON-RPC error for that slot. Streaming into a batch response would require
// multiple frames for a single logical JSON array, which violates JSON-RPC.
func TestStreamableResult_WebSocketBatchReturnsError(t *testing.T) {
	t.Parallel()

	s := newWSStreamingServer()
	defer s.Close()

	d := websocket.Dialer{}
	c, _, err := d.Dial("ws://"+s.Listener.Addr().String()+"/websocket", nil)
	require.NoError(t, err)
	defer c.Close()

	// Two-element batch: one streamable + one plain, so the response is always
	// an array regardless of the single-element serialisation quirk.
	batch := `[{"jsonrpc":"2.0","method":"stream","id":"1","params":{}},{"jsonrpc":"2.0","method":"echo","id":"2","params":{}}]`
	err = c.WriteMessage(websocket.TextMessage, []byte(batch))
	require.NoError(t, err)

	_, msg, err := c.ReadMessage()
	require.NoError(t, err)

	var resps []map[string]any
	require.NoError(t, json.Unmarshal(msg, &resps))
	require.Len(t, resps, 2,
		"batch response must have one slot per request (got %s)", msg)

	// Find the streamable slot by id.
	var streamSlot map[string]any
	for _, r := range resps {
		if r["id"] == "1" {
			streamSlot = r
		}
	}
	require.NotNil(t, streamSlot, "streamable slot must be present")
	require.Contains(t, streamSlot, "error",
		"batch slot with streamable result must carry a JSON-RPC error (got %v)", streamSlot)
}
