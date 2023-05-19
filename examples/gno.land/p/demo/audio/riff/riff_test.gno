package riff

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func TestRiff(t *testing.T) {
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	w, err := NewWriter(foo, []byte("WAVE"), 243800)
	if err != nil {
		t.Errorf("%s", err)
	}
	_, err = w.WriteChunk([]byte("TEST"), 18)
	if err != nil {
		t.Errorf("%s", err)
	}
	foo.Flush()
	out := fmt.Sprintf("%x", b.Bytes())
	if out != "5249464658b80300574156455445535412000000" {
		t.Errorf("got: %s", out)
	}
}
