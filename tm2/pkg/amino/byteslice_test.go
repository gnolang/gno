package amino

import (
	"bytes"
	"testing"
)

func TestReadByteSliceEquality(t *testing.T) {
	t.Parallel()

	var encoded []byte
	var err error
	cdc := NewCodec()
	type byteWrapper struct {
		Val []byte
	}
	// Write a byteslice
	testBytes := byteWrapper{[]byte("ThisIsSomeTestArrayEmbeddedInAStruct")}
	encoded, err = cdc.MarshalSized(testBytes)
	if err != nil {
		t.Error(err.Error())
	}

	// Read the byteslice, should return the same byteslice
	var testBytes2 byteWrapper
	err = cdc.UnmarshalSized(encoded, &testBytes2)
	if err != nil {
		t.Error(err.Error())
	}

	if !bytes.Equal(testBytes.Val, testBytes2.Val) {
		t.Error("Returned the wrong bytes")
	}
}
