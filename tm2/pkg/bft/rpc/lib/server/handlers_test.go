package rpcserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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

	rs.RegisterRPCFuncs(mux, funcMap, log.NewNoopLogger(), 0)

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

// streamCalls counts the calls of the "stream" method of streamingTestMux.
var streamCalls atomic.Int64

func streamingTestMux() *http.ServeMux {
	funcMap := map[string]*rs.RPCFunc{
		"stream": rs.NewRPCFunc(func(ctx *types.Context) (*streamableTestResult, error) {
			streamCalls.Add(1)
			return &streamableTestResult{Greeting: "hello"}, nil
		}, ""),
	}
	mux := http.NewServeMux()
	rs.RegisterRPCFuncs(mux, funcMap, log.NewNoopLogger(), 0)
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

// TestStreamableResult_HeadDoesNotStream checks that a HEAD request gets the
// status and headers of a streamed result and no body, on the route and on the
// JSON-RPC path, whose call a HEAD request carries in its body.
func TestStreamableResult_HeadDoesNotStream(t *testing.T) {
	mux := streamingTestMux()
	call := `{"jsonrpc":"2.0","method":"stream","id":"1"}`

	for _, path := range []string{"/stream", "/"} {
		req := httptest.NewRequest(http.MethodHead, path, strings.NewReader(call))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), path)
		assert.Zero(t, rec.Body.Len(), "HEAD %s streamed %q", path, rec.Body)
	}
}

// TestJSONRPCBatch_HeadRunsNoCall checks that a HEAD batch gets the status and
// headers before its first call, while every other method runs the batch and
// a single HEAD call still runs.
func TestJSONRPCBatch_HeadRunsNoCall(t *testing.T) {
	mux := streamingTestMux()
	batch := `[{"jsonrpc":"2.0","method":"stream","id":"1"},{"jsonrpc":"2.0","method":"stream","id":"2"}]`

	for _, body := range []string{batch, " \n" + batch} {
		streamCalls.Store(0)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, "/", strings.NewReader(body)))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Zero(t, rec.Body.Len(), "HEAD batch answered %q", rec.Body)
		assert.Zero(t, streamCalls.Load(), "HEAD batch %q ran a call", body)
	}

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		streamCalls.Store(0)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, "/", strings.NewReader(batch)))

		assert.NotZero(t, rec.Body.Len(), method)
		assert.EqualValues(t, 2, streamCalls.Load(), method)
	}

	streamCalls.Store(0)
	single := `{"jsonrpc":"2.0","method":"stream","id":"1"}`
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodHead, "/", strings.NewReader(single)))
	assert.EqualValues(t, 1, streamCalls.Load())
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

// batchOf returns a JSON-RPC batch of n calls to method "c".
func batchOf(n int) string {
	return batchCalls(slices.Repeat([]string{"c"}, n)...)
}

// TestRPCBatchLimit verifies that a batch over the limit is refused whole,
// with one error and no call run, and that a batch at the limit is served.
func TestRPCBatchLimit(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64
	funcMap := map[string]*rs.RPCFunc{
		"c": rs.NewRPCFunc(func(*types.Context) (string, error) {
			calls.Add(1)
			return "foo", nil
		}, ""),
	}
	mux := http.NewServeMux()
	rs.RegisterRPCFuncs(mux, funcMap, log.NewNoopLogger(), 2)

	post := func(payload string) []byte {
		t.Helper()
		req, _ := http.NewRequest("POST", "http://localhost/", strings.NewReader(payload))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		blob, err := io.ReadAll(rec.Result().Body)
		require.NoError(t, err)
		return blob
	}

	var responses types.RPCResponses
	require.NoError(t, json.Unmarshal(post(batchOf(2)), &responses))
	require.Len(t, responses, 2)
	for _, resp := range responses {
		require.Nil(t, resp.Error)
	}
	require.Equal(t, int64(2), calls.Load())

	requireBatchRefused(t, post(batchOf(3)))
	assert.Equal(t, int64(2), calls.Load(), "no call of an over-limit batch may run")
}

// requireBatchRefused checks that blob is the one error object a batch of 3
// requests gets from a server whose limit is 2.
func requireBatchRefused(t *testing.T, blob []byte) {
	t.Helper()
	var resp types.RPCResponse
	require.NoError(t, json.Unmarshal(blob, &resp), "an over-limit batch gets one response object, got %s", blob)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32600, resp.Error.Code)
	assert.Contains(t, resp.Error.Data, "batch of 3 requests exceeds the limit of 2")
}

// batchTestFuncs returns the methods the batch streaming tests call: "big"
// answers 64 KiB, "gated" answers once release is closed or fails after 5 s,
// and "panics" panics. Every call of "big" and "gated" adds one to calls.
func batchTestFuncs(calls *atomic.Int64, release <-chan struct{}) map[string]*rs.RPCFunc {
	big := strings.Repeat("x", 64<<10)
	return map[string]*rs.RPCFunc{
		"big": rs.NewRPCFunc(func(*types.Context) (string, error) {
			calls.Add(1)
			return big, nil
		}, ""),
		"gated": rs.NewRPCFunc(func(*types.Context) (string, error) {
			calls.Add(1)
			select {
			case <-release:
				return big, nil
			case <-time.After(5 * time.Second):
				return "", errors.New("not released within 5s")
			}
		}, ""),
		"panics": rs.NewRPCFunc(func(*types.Context) (string, error) {
			panic("call failed")
		}, ""),
		"small": rs.NewRPCFunc(func(*types.Context) (string, error) {
			calls.Add(1)
			time.Sleep(time.Millisecond)
			return "ok", nil
		}, ""),
	}
}

// batchCalls returns a batch calling each method in turn, with ids "0" up.
func batchCalls(methods ...string) string {
	calls := make([]string, len(methods))
	for i, m := range methods {
		calls[i] = fmt.Sprintf(`{"jsonrpc":"2.0","method":%q,"id":"%d"}`, m, i)
	}
	return "[" + strings.Join(calls, ",") + "]"
}

// gatedBatch returns a batch of two "big" calls followed by n-2 "gated"
// ones. WebSocket holds a first answer until a second one makes the message
// an array, so two answers are written before the first gated call.
func gatedBatch(n int) string {
	return batchCalls(append([]string{"big", "big"}, slices.Repeat([]string{"gated"}, n-2)...)...)
}

// batchServer serves batchTestFuncs over HTTP and over WebSocket at
// /websocket, for one request; served is closed once its handler returns.
func batchServer(release <-chan struct{}) (s *httptest.Server, calls *atomic.Int64, served chan struct{}) {
	calls, served = new(atomic.Int64), make(chan struct{})
	funcs := batchTestFuncs(calls, release)
	wm := rs.NewWebsocketManager(funcs)
	wm.SetLogger(log.NewNoopLogger())
	mux := http.NewServeMux()
	rs.RegisterRPCFuncs(mux, funcs, log.NewNoopLogger(), 0)
	mux.HandleFunc("/websocket", wm.WebsocketHandler)
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(served)
		mux.ServeHTTP(w, r)
	}))
	return s, calls, served
}

// requireStreamed reads the first bytes of the answer to gatedBatch(3) from r,
// then releases the gated call and checks that it answered without error.
func requireStreamed(t *testing.T, r io.Reader, release chan struct{}) {
	t.Helper()
	first := make([]byte, 100)
	_, err := io.ReadFull(r, first)
	require.NoError(t, err)
	close(release)
	rest, err := io.ReadAll(r)
	require.NoError(t, err)

	var responses types.RPCResponses
	require.NoError(t, json.Unmarshal(append(first, rest...), &responses))
	require.Len(t, responses, 3)
	require.Nil(t, responses[2].Error, "the first answer reached the client only after the last call returned")
}

// requireStopped releases the gated calls of a batch of n whose client has
// gone, and checks that the batch ends before running them all. Over
// WebSocket the handler returns with writeRoutine while readRoutine may still
// be running calls, so the count is watched for a second after that.
func requireStopped(t *testing.T, release chan struct{}, served <-chan struct{}, calls *atomic.Int64, n int) {
	t.Helper()
	close(release)
	select {
	case <-served:
	case <-time.After(10 * time.Second):
		t.Fatal("the batch was still running 10s after the client had gone")
	}
	assert.Never(t, func() bool { return calls.Load() >= int64(n) }, time.Second, 10*time.Millisecond,
		"every call ran after the client had gone")
}

// TestRPCBatchStreamsAnswers verifies that the first answer of a batch
// reaches the client while a later call of the batch is still running.
func TestRPCBatchStreamsAnswers(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	s, _, _ := batchServer(release)
	defer s.Close()

	res, err := http.Post(s.URL, "application/json", strings.NewReader(gatedBatch(3)))
	require.NoError(t, err)
	defer res.Body.Close()
	requireStreamed(t, res.Body, release)
}

// TestRPCBatchStopsOnWriteError verifies that a batch runs no further call
// once the client has gone and writing an answer fails.
func TestRPCBatchStopsOnWriteError(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	s, calls, served := batchServer(release)
	defer s.Close()

	const n = 50
	res, err := http.Post(s.URL, "application/json", strings.NewReader(gatedBatch(n)))
	require.NoError(t, err)
	_, err = io.ReadFull(res.Body, make([]byte, 100))
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	requireStopped(t, release, served, calls, n)
}

// TestRPCBatchPanicAnswersInBand verifies that a call panicking inside a batch
// gets an internal error in its own slot and the other calls their answers.
func TestRPCBatchPanicAnswersInBand(t *testing.T) {
	t.Parallel()

	s, _, _ := batchServer(nil)
	defer s.Close()

	res, err := http.Post(s.URL, "application/json", strings.NewReader(batchCalls("big", "panics", "big")))
	require.NoError(t, err)
	defer res.Body.Close()
	blob, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	requirePanicInBand(t, blob)
}

// requirePanicInBand checks that blob answers the batch "big", "panics",
// "big" with the panic as an internal error in the second slot.
func requirePanicInBand(t *testing.T, blob []byte) {
	t.Helper()
	var responses types.RPCResponses
	require.NoError(t, json.Unmarshal(blob, &responses), "got %.200s", blob)
	require.Len(t, responses, 3)
	assert.Nil(t, responses[0].Error)
	assert.Nil(t, responses[2].Error)
	require.NotNil(t, responses[1].Error)
	assert.Equal(t, types.JSONRPCStringID("1"), responses[1].ID)
	assert.Equal(t, -32603, responses[1].Error.Code)
	assert.Equal(t, "call failed", responses[1].Error.Data)
}

// TestRPCBatchStopsOnHangUpWithSmallAnswers verifies that a batch of small
// answers stops within a few calls of its client hanging up, rather than once
// net/http's buffers fill.
func TestRPCBatchStopsOnHangUpWithSmallAnswers(t *testing.T) {
	t.Parallel()

	s, calls, served := batchServer(nil)
	defer s.Close()

	body := batchCalls(slices.Repeat([]string{"small"}, 1000)...)
	c, err := net.Dial("tcp", s.Listener.Addr().String())
	require.NoError(t, err)
	_, err = fmt.Fprintf(c, "POST / HTTP/1.1\r\nHost: x\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	require.NoError(t, err)
	require.Eventually(t, func() bool { return calls.Load() >= 10 }, 5*time.Second, time.Millisecond)
	atHangUp := calls.Load()
	require.NoError(t, c.Close())

	select {
	case <-served:
	case <-time.After(10 * time.Second):
		t.Fatal("the batch was still running 10s after the client had gone")
	}
	// Unflushed, about 60 of these answers fit in net/http's 4 KiB buffer,
	// and the batch ran over 100 more calls.
	assert.Less(t, calls.Load()-atHangUp, int64(10), "calls run after the client hung up")
}

// nullBatch is a batch of two "big" calls around a null element.
const nullBatch = `[{"jsonrpc":"2.0","method":"big","id":"0"},null,{"jsonrpc":"2.0","method":"big","id":"2"}]`

// TestRPCBatchNullElement verifies that a null element of a batch gets an
// invalid-request error in its own slot and the other calls their answers.
func TestRPCBatchNullElement(t *testing.T) {
	t.Parallel()

	s, _, _ := batchServer(nil)
	defer s.Close()

	res, err := http.Post(s.URL, "application/json", strings.NewReader(nullBatch))
	require.NoError(t, err)
	defer res.Body.Close()
	blob, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
	requireNullInBand(t, blob)
}

// requireNullInBand checks that blob answers nullBatch in order: a result, an
// invalid-request error in the null element's place, and a result. The error's
// id is empty, so its place is all that ties it to the element.
func requireNullInBand(t *testing.T, blob []byte) {
	t.Helper()
	var responses types.RPCResponses
	require.NoError(t, json.Unmarshal(blob, &responses), "got %.200s", blob)
	require.Len(t, responses, 3)
	assert.Nil(t, responses[0].Error)
	assert.Equal(t, types.JSONRPCStringID("0"), responses[0].ID)
	require.NotNil(t, responses[1].Error)
	assert.Equal(t, -32600, responses[1].Error.Code)
	assert.Nil(t, responses[2].Error)
	assert.Equal(t, types.JSONRPCStringID("2"), responses[2].ID)
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

// TestWebsocketBatchLimit verifies that a batch over the limit is refused
// whole, with one error and no call run, and that the connection keeps
// serving requests afterwards.
func TestWebsocketBatchLimit(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64
	funcMap := map[string]*rs.RPCFunc{
		"c": rs.NewWSRPCFunc(func(*types.Context) (string, error) {
			calls.Add(1)
			return "foo", nil
		}, ""),
	}
	wm := rs.NewWebsocketManager(funcMap, rs.MaxBatchRequests(2))
	wm.SetLogger(log.NewNoopLogger())
	mux := http.NewServeMux()
	mux.HandleFunc("/websocket", wm.WebsocketHandler)
	s := httptest.NewServer(mux)
	defer s.Close()

	c := dialBatchServer(t, s)
	defer c.Close()

	require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(batchOf(3))))
	_, msg, err := c.ReadMessage()
	require.NoError(t, err)
	requireBatchRefused(t, msg)
	assert.Equal(t, int64(0), calls.Load(), "no call of an over-limit batch may run")

	require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(batchOf(2))))
	_, msg, err = c.ReadMessage()
	require.NoError(t, err)
	var responses types.RPCResponses
	require.NoError(t, json.Unmarshal(msg, &responses), "got %s", msg)
	require.Len(t, responses, 2)
	for _, resp := range responses {
		require.Nil(t, resp.Error)
	}
	assert.Equal(t, int64(2), calls.Load())
}

// dialBatchServer opens a WebSocket connection to s at /websocket.
func dialBatchServer(t *testing.T, s *httptest.Server) *websocket.Conn {
	t.Helper()
	c, _, err := websocket.DefaultDialer.Dial("ws://"+s.Listener.Addr().String()+"/websocket", nil)
	require.NoError(t, err)
	return c
}

// TestWebsocketBatchStreamsAnswers verifies that the first answer of a batch
// reaches the client while a later call is still running, inside the one
// message the batch is answered with, and that a batch with one answer still
// gets one object.
func TestWebsocketBatchStreamsAnswers(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	s, _, _ := batchServer(release)
	defer s.Close()
	c := dialBatchServer(t, s)
	defer c.Close()

	require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(batchCalls("big"))))
	_, msg, err := c.ReadMessage()
	require.NoError(t, err)
	var one types.RPCResponse
	require.NoError(t, json.Unmarshal(msg, &one), "a batch with one answer gets one object, got %.100s", msg)

	require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(gatedBatch(3))))
	_, r, err := c.NextReader()
	require.NoError(t, err)
	requireStreamed(t, r, release)
}

// readRoutines counts the goroutines of this process running a WebSocket
// connection's read loop.
func readRoutines() int {
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	return strings.Count(string(buf[:n]), "(*wsConnection).readRoutine")
}

// TestWebsocketBatchStopsOnWriteError verifies that a batch runs no further
// call once the client has gone and writing an answer fails, and that the
// connection's read loop then exits. It counts read loops process-wide, so it
// does not run in parallel.
func TestWebsocketBatchStopsOnWriteError(t *testing.T) {
	before := readRoutines()
	release := make(chan struct{})
	s, calls, served := batchServer(release)
	defer s.Close()
	c := dialBatchServer(t, s)

	const n = 50
	require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(gatedBatch(n))))
	_, r, err := c.NextReader()
	require.NoError(t, err)
	_, err = io.ReadFull(r, make([]byte, 100))
	require.NoError(t, err)
	require.NoError(t, c.Close())
	requireStopped(t, release, served, calls, n)
	require.Eventually(t, func() bool { return readRoutines() <= before }, 3*time.Second, 10*time.Millisecond,
		"the read loop still runs after its connection stopped")
}

// TestWebsocketBatchWriteDeadlinePerAnswer verifies that the write deadline
// is set again before each answer of a batch, so a batch whose calls outlast
// one deadline still arrives whole.
func TestWebsocketBatchWriteDeadlinePerAnswer(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	wm := rs.NewWebsocketManager(batchTestFuncs(new(atomic.Int64), release), rs.WriteWait(200*time.Millisecond))
	wm.SetLogger(log.NewNoopLogger())
	s := httptest.NewServer(http.HandlerFunc(wm.WebsocketHandler))
	defer s.Close()
	c, _, err := websocket.DefaultDialer.Dial("ws://"+s.Listener.Addr().String(), nil)
	require.NoError(t, err)
	defer c.Close()

	// The first answer sets a deadline that expires while the gated call waits.
	require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(batchCalls("big"))))
	_, _, err = c.ReadMessage()
	require.NoError(t, err)
	require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(gatedBatch(3))))
	time.AfterFunc(500*time.Millisecond, func() { close(release) })
	_, msg, err := c.ReadMessage()
	require.NoError(t, err)
	var responses types.RPCResponses
	require.NoError(t, json.Unmarshal(msg, &responses))
	require.Len(t, responses, 3)
}

// TestWebsocketBatchPanicAnswersInBand verifies that a call panicking inside
// a batch gets an internal error in its own slot and the other calls their
// answers.
func TestWebsocketBatchPanicAnswersInBand(t *testing.T) {
	t.Parallel()

	s, _, _ := batchServer(nil)
	defer s.Close()
	c := dialBatchServer(t, s)
	defer c.Close()

	require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(batchCalls("big", "panics", "big"))))
	_, msg, err := c.ReadMessage()
	require.NoError(t, err)
	requirePanicInBand(t, msg)
}

// TestWebsocketBatchNullElement verifies that a null element of a batch gets
// an invalid-request error and the other calls their answers.
func TestWebsocketBatchNullElement(t *testing.T) {
	t.Parallel()

	s, _, _ := batchServer(nil)
	defer s.Close()
	c := dialBatchServer(t, s)
	defer c.Close()

	require.NoError(t, c.WriteMessage(websocket.TextMessage, []byte(nullBatch)))
	_, msg, err := c.ReadMessage()
	require.NoError(t, err)
	requireNullInBand(t, msg)
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
