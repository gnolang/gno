package amino

import (
	"bytes"
	"testing"
)

// failNthWriter fails only its nth Write, then accepts again.
type failNthWriter struct{ n, calls int }

func (w *failNthWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == w.n {
		return 0, bytes.ErrTooLarge
	}
	return len(p), nil
}

// A write error inside a list is returned, not masked by the closing bracket.
func TestMarshalJSONToListWriteError(t *testing.T) {
	v := []string{"a", "bb", "ccc"}
	for n := 1; ; n++ {
		w := &failNthWriter{n: n}
		err := NewCodec().MarshalJSONTo(w, v) // not gcdc: marshaling seals it
		// Every write succeeded.
		if w.calls < n {
			if err != nil {
				t.Fatalf("n=%d: unexpected error %v", n, err)
			}
			return
		}
		if err == nil {
			t.Fatalf("n=%d: write %d failed but MarshalJSONTo returned nil", n, n)
		}
	}
}

// Verify that the output is the same whether the input is
// written all at once or one byte at a time.
func TestIfaceJSONWriterChunking(t *testing.T) {
	tests := []struct {
		name   string
		object bool
		in     string
		want   string
	}{
		{"empty object", true, `{}`, `}`},
		{"one field", true, `{"A":"one"}`, `,"A":"one"}`},
		{"two fields", true, `{"Name":"glider","MaxAltitude":"42"}`, `,"Name":"glider","MaxAltitude":"42"}`},
		{"value string", false, `"42"`, `"42"`},
		{"value list", false, `[1,2]`, `[1,2]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var whole, chunked bytes.Buffer

			w1 := ifaceJSONWriter{w: &whole, object: tt.object}
			if _, err := w1.Write([]byte(tt.in)); err != nil {
				t.Fatalf("whole write: %v", err)
			}
			if err := w1.finish(); err != nil {
				t.Fatalf("whole finish: %v", err)
			}

			w2 := ifaceJSONWriter{w: &chunked, object: tt.object}
			for i := range tt.in {
				if _, err := w2.Write([]byte(tt.in[i : i+1])); err != nil {
					t.Fatalf("byte %d: %v", i, err)
				}
			}
			if err := w2.finish(); err != nil {
				t.Fatalf("chunked finish: %v", err)
			}

			if whole.String() != tt.want {
				t.Errorf("whole write: got %q, want %q", whole.String(), tt.want)
			}
			if chunked.String() != whole.String() {
				t.Errorf("chunked: got %q, want %q", chunked.String(), whole.String())
			}
		})
	}
}
