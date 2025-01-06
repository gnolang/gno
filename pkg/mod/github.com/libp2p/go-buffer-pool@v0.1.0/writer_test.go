package pool

import (
	"bytes"
	"testing"
)

func checkSize(t *testing.T, w *Writer) {
	if w.Size()-w.Buffered() != w.Available() {
		t.Fatalf("size (%d), buffered (%d), available (%d) mismatch", w.Size(), w.Buffered(), w.Available())
	}
}

func TestWriter(t *testing.T) {
	var b bytes.Buffer
	w := Writer{W: &b}
	n, err := w.Write([]byte("foobar"))
	checkSize(t, &w)

	if err != nil || n != 6 {
		t.Fatalf("write failed: %d, %s", n, err)
	}
	if b.Len() != 0 {
		t.Fatal("expected the buffer to be empty")
	}
	if w.Buffered() != 6 {
		t.Fatalf("expected 6 bytes to be buffered, got %d", w.Buffered())
	}
	checkSize(t, &w)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	checkSize(t, &w)
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	checkSize(t, &w)
	if b.String() != "foobar" {
		t.Fatal("expected to have written foobar")
	}
	b.Reset()

	buf := make([]byte, WriterBufferSize)
	n, err = w.Write(buf)
	if n != WriterBufferSize || err != nil {
		t.Fatalf("write failed: %d, %s", n, err)
	}
	checkSize(t, &w)
	if b.Len() != WriterBufferSize {
		t.Fatal("large write should have gone through directly")
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	checkSize(t, &w)

	b.Reset()
	if err := w.WriteByte(1); err != nil {
		t.Fatal(err)
	}
	if w.Buffered() != 1 {
		t.Fatalf("expected 1 byte to be buffered, got %d", w.Buffered())
	}
	if n, err := w.WriteRune('1'); err != nil || n != 1 {
		t.Fatal(err)
	}
	if w.Buffered() != 2 {
		t.Fatalf("expected 2 bytes to be buffered, got %d", w.Buffered())
	}
	checkSize(t, &w)
	if n, err := w.WriteString("foobar"); err != nil || n != 6 {
		t.Fatal(err)
	}
	if w.Buffered() != 8 {
		t.Fatalf("expected 8 bytes to be buffered, got %d", w.Buffered())
	}
	checkSize(t, &w)
	if b.Len() != 0 {
		t.Fatal("write should have been buffered")
	}
	n, err = w.Write(buf)
	if n != WriterBufferSize || err != nil {
		t.Fatalf("write failed: %d, %s", n, err)
	}
	if b.Len() != WriterBufferSize || b.Bytes()[0] != 1 || b.String()[1:8] != "1foobar" {
		t.Fatalf("failed to flush properly: len:%d, prefix:%#v", b.Len(), b.Bytes()[:10])
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}
