package rpcserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// failingResponseWriter is an http.ResponseWriter whose Write always fails
// with the configured error. Header and WriteHeader are no-ops that record
// the last status code. Used to verify the response-write path returns the
// underlying error instead of panicking when a client disconnects.
type failingResponseWriter struct {
	header     http.Header
	writeErr   error
	statusCode int
}

func newFailingResponseWriter(writeErr error) *failingResponseWriter {
	return &failingResponseWriter{
		header:   http.Header{},
		writeErr: writeErr,
	}
}

func (w *failingResponseWriter) Header() http.Header        { return w.header }
func (w *failingResponseWriter) WriteHeader(statusCode int) { w.statusCode = statusCode }
func (w *failingResponseWriter) Write(_ []byte) (int, error) {
	return 0, w.writeErr
}

// validRPCResponse builds a response directly via struct literal to avoid
// NewRPCSuccessResponse, which routes through amino's marshaling and panics
// on unregistered types.
func validRPCResponse() types.RPCResponse {
	return types.RPCResponse{
		JSONRPC: "2.0",
		ID:      types.JSONRPCStringID("1"),
		Result:  json.RawMessage(`{"ok":"yes"}`),
	}
}

func TestWriteRPCResponseHTTP_WriteErrorReturned(t *testing.T) {
	w := newFailingResponseWriter(io.ErrClosedPipe)

	require.NotPanics(t, func() {
		err := WriteRPCResponseHTTP(w, validRPCResponse())
		require.Error(t, err)
		assert.True(t, errors.Is(err, io.ErrClosedPipe),
			"expected returned error to wrap io.ErrClosedPipe, got %v", err)
	})
}

func TestWriteRPCResponseHTTP_SuccessReturnsNil(t *testing.T) {
	rec := httptest.NewRecorder()

	err := WriteRPCResponseHTTP(rec, validRPCResponse())
	require.NoError(t, err)

	res := rec.Result()
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"ok"`)
	assert.Contains(t, string(body), `"yes"`)
}

func TestWriteRPCResponseHTTPError_WriteErrorReturned(t *testing.T) {
	w := newFailingResponseWriter(io.ErrClosedPipe)

	require.NotPanics(t, func() {
		err := WriteRPCResponseHTTPError(w, http.StatusInternalServerError, validRPCResponse())
		require.Error(t, err)
		assert.True(t, errors.Is(err, io.ErrClosedPipe),
			"expected returned error to wrap io.ErrClosedPipe, got %v", err)
	})
}

func TestWriteRPCResponseHTTPError_SuccessReturnsNil(t *testing.T) {
	rec := httptest.NewRecorder()

	err := WriteRPCResponseHTTPError(rec, http.StatusInternalServerError, validRPCResponse())
	require.NoError(t, err)

	res := rec.Result()
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

// TestBatchWriterMatchesMarshalIndent verifies that batchWriter writes every
// prefix of a varied list of answers in the bytes json.MarshalIndent gives the
// whole slice.
func TestBatchWriterMatchesMarshalIndent(t *testing.T) {
	answers := types.RPCResponses{
		validRPCResponse(),
		{JSONRPC: "2.0", ID: types.JSONRPCIntID(7), Result: json.RawMessage(`{"a":[1,{"b":[]},{}],"c":"<&> é"}`)},
		types.RPCInternalError(types.JSONRPCStringID(`"quoted" <id>`), errors.New("line\nbreak")),
		types.RPCMethodNotFoundError(types.JSONRPCStringID("")),
		{JSONRPC: "2.0", ID: types.JSONRPCStringID("s"), Result: json.RawMessage(` [ "x" , null ] `)},
		{JSONRPC: "2.0", ID: types.JSONRPCStringID("n")},
	}
	for n := 1; n <= len(answers); n++ {
		var got bytes.Buffer
		b := batchWriter{w: &got}
		for _, resp := range answers[:n] {
			require.NoError(t, b.write(resp))
		}
		require.NoError(t, b.close())

		want, err := json.MarshalIndent(answers[:n], "", "  ")
		require.NoError(t, err)
		require.Equal(t, string(want), got.String(), "first %d answers", n)
	}
}

// TestBatchWriterReplacesUnmarshalableAnswer verifies that an answer that does
// not marshal becomes an internal error for its id, keeping the array valid.
func TestBatchWriterReplacesUnmarshalableAnswer(t *testing.T) {
	var got bytes.Buffer
	b := batchWriter{w: &got}
	require.NoError(t, b.write(types.RPCResponse{JSONRPC: "2.0", ID: types.JSONRPCStringID("1"), Result: json.RawMessage(`{`)}))
	require.NoError(t, b.close())

	var responses types.RPCResponses
	require.NoError(t, json.Unmarshal(got.Bytes(), &responses), "got %s", got.Bytes())
	require.Len(t, responses, 1)
	require.NotNil(t, responses[0].Error)
	assert.Equal(t, types.JSONRPCStringID("1"), responses[0].ID)
	assert.Equal(t, -32603, responses[0].Error.Code)
}

// streamingResult is a test StreamableResult that writes the configured chunks
// directly to the writer, recording any write error returned.
type streamingResult struct {
	chunks  [][]byte
	written int
}

func (s *streamingResult) StreamJSON(_ context.Context, w io.Writer) error {
	for _, c := range s.chunks {
		n, err := w.Write(c)
		s.written += n
		if err != nil {
			return err
		}
	}
	return nil
}

// TestWriteStreamingRPCResponseHTTP_WritesEnvelope verifies that a streaming
// response writes the JSON-RPC envelope around the StreamJSON output without
// buffering the result body in memory.
func TestWriteStreamingRPCResponseHTTP_WritesEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	body := &streamingResult{chunks: [][]byte{
		[]byte(`{"chunk_a":"first",`),
		[]byte(`"chunk_b":"second"}`),
	}}

	err := WriteStreamingRPCResponseHTTP(httptest.NewRequest(http.MethodGet, "/", nil), rec, types.JSONRPCStringID("1"), body)
	require.NoError(t, err)

	res := rec.Result()
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	got, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	// The envelope must be valid JSON parseable as a full RPCResponse, with
	// the result chunks reassembled into the result field.
	var parsed struct {
		JSONRPC string         `json:"jsonrpc"`
		ID      string         `json:"id"`
		Result  map[string]any `json:"result"`
	}
	require.NoError(t, json.Unmarshal(got, &parsed))
	assert.Equal(t, "2.0", parsed.JSONRPC)
	assert.Equal(t, "1", parsed.ID)
	assert.Equal(t, "first", parsed.Result["chunk_a"])
	assert.Equal(t, "second", parsed.Result["chunk_b"])
}

// TestWriteStreamingRPCResponseHTTP_PropagatesWriteError verifies that a write
// error during streaming is returned (not panicked) — same contract as the
// non-streaming response writers.
func TestWriteStreamingRPCResponseHTTP_PropagatesWriteError(t *testing.T) {
	w := newFailingResponseWriter(io.ErrClosedPipe)
	body := &streamingResult{chunks: [][]byte{[]byte(`{"x":1}`)}}

	require.NotPanics(t, func() {
		err := WriteStreamingRPCResponseHTTP(httptest.NewRequest(http.MethodGet, "/", nil), w, types.JSONRPCStringID("1"), body)
		require.Error(t, err)
		assert.True(t, errors.Is(err, io.ErrClosedPipe),
			"expected returned error to wrap io.ErrClosedPipe, got %v", err)
	})
}

// slowStreamingResult writes its chunks with a pause between them, so the
// response takes longer to produce than the server's WriteTimeout allows.
type slowStreamingResult struct {
	chunks [][]byte
	pause  time.Duration
}

func (s *slowStreamingResult) StreamJSON(_ context.Context, w io.Writer) error {
	for i, c := range s.chunks {
		if i > 0 {
			time.Sleep(s.pause)
		}
		if _, err := w.Write(c); err != nil {
			return err
		}
	}
	return nil
}

// A response that makes progress may outlive the original deadline.
func TestStreamingResponseOutlivesWriteTimeout(t *testing.T) {
	t.Parallel()

	const (
		writeTimeout = 300 * time.Millisecond
		chunks       = 6
		pause        = 100 * time.Millisecond
	)

	body := &slowStreamingResult{pause: pause}
	for i := range chunks {
		body.chunks = append(body.chunks, fmt.Appendf(nil, `"chunk-%d",`, i))
	}
	// Valid JSON once the envelope closes it: an array of the chunks above.
	body.chunks = append([][]byte{[]byte(`[`)}, body.chunks...)
	body.chunks = append(body.chunks, []byte(`"last"]`))

	config := DefaultConfig()
	config.WriteTimeout = writeTimeout

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := WriteStreamingRPCResponseHTTP(r, w, types.JSONRPCStringID("1"), body); err != nil {
			t.Errorf("stream: %v", err)
		}
	})

	l, err := Listen("tcp://127.0.0.1:0", config)
	require.NoError(t, err)
	t.Cleanup(func() { l.Close() })
	go StartHTTPServer(l, mux, log.NewTestingLogger(t), config)

	// The whole response takes chunks*pause, comfortably past writeTimeout.
	started := time.Now()
	res, err := http.Get("http://" + l.Addr().String() + "/")
	require.NoError(t, err)
	defer res.Body.Close()

	got, err := io.ReadAll(res.Body)
	require.NoError(t, err, "the body was cut before it finished")
	elapsed := time.Since(started)
	require.Greater(t, elapsed, writeTimeout,
		"the response finished before WriteTimeout, so it never tested anything")

	// Parseable end to end is the assertion that matters: a truncated stream
	// still arrives with a 200 and a plausible prefix, and only the parse
	// distinguishes the two.
	var parsed struct {
		JSONRPC string   `json:"jsonrpc"`
		Result  []string `json:"result"`
	}
	require.NoError(t, json.Unmarshal(got, &parsed),
		"response is not valid JSON, so it was truncated: %q", got)
	assert.Equal(t, "2.0", parsed.JSONRPC)
	require.Len(t, parsed.Result, chunks+1)
	assert.Equal(t, "last", parsed.Result[chunks])
}

type deadlineRecorder struct {
	*httptest.ResponseRecorder
	deadlines []time.Time
}

func (r *deadlineRecorder) SetWriteDeadline(deadline time.Time) error {
	r.deadlines = append(r.deadlines, deadline)
	return nil
}

func TestProgressDeadlineWriterKeepsFiniteDeadline(t *testing.T) {
	rec := &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
	wrapped := &ResponseWriterWrapper{ResponseWriter: rec}
	w := &progressDeadlineWriter{
		w:       wrapped,
		rc:      http.NewResponseController(wrapped),
		timeout: time.Minute,
	}

	_, err := w.Write([]byte("first"))
	require.NoError(t, err)
	require.Len(t, rec.deadlines, 1)
	require.True(t, rec.deadlines[0].After(time.Now()))

	_, err = w.Write([]byte("second"))
	require.NoError(t, err)
	require.Len(t, rec.deadlines, 1)

	w.rearm = time.Now().Add(-time.Second)
	_, err = w.Write([]byte("third"))
	require.NoError(t, err)
	require.Len(t, rec.deadlines, 2)
	require.True(t, rec.deadlines[1].After(rec.deadlines[0]))
	require.Equal(t, "firstsecondthird", rec.Body.String())
}
