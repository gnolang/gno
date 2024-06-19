package amino_test

import (
	"bytes"
	"testing"
	"time"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	type SimpleStruct struct {
		String string
		Bytes  []byte
		Time   time.Time
	}

	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().UTC().Truncate(time.Millisecond), // strip monotonic and timezone.
	}

	b, err := cdc.MarshalSized(s)
	assert.Nil(t, err)
	t.Logf("MarshalSized(s) -> %X", b)

	var s2 SimpleStruct
	err = cdc.UnmarshalSized(b, &s2)
	assert.Nil(t, err)
	assert.Equal(t, s, s2)
}

func TestUnmarshalReader(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	type SimpleStruct struct {
		String string
		Bytes  []byte
		Time   time.Time
	}

	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().UTC().Truncate(time.Millisecond), // strip monotonic and timezone.
	}

	b, err := cdc.MarshalSized(s)
	assert.Nil(t, err)
	t.Logf("MarshalSized(s) -> %X", b)

	var s2 SimpleStruct
	_, err = cdc.UnmarshalSizedReader(bytes.NewBuffer(b), &s2, 0)
	assert.Nil(t, err)

	assert.Equal(t, s, s2)
}

type stringWrapper struct {
	S string
}

func TestUnmarshalReaderSize(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	s1 := stringWrapper{"foo"}
	b, err := cdc.MarshalSized(s1)
	assert.Nil(t, err)
	t.Logf("MarshalSized(s) -> %X", b)

	var s2 stringWrapper
	var n int64
	n, err = cdc.UnmarshalSizedReader(bytes.NewBuffer(b), &s2, 0)
	assert.Nil(t, err)
	assert.Equal(t, s1, s2)
	frameLengthBytes, msgLengthBytes, embedOverhead := 1, 1, 1
	assert.Equal(t, frameLengthBytes+msgLengthBytes+embedOverhead+len(s1.S), int(n))
}

func TestUnmarshalReaderSizeLimit(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	s1 := stringWrapper{"foo"}
	b, err := cdc.MarshalSized(s1)
	assert.Nil(t, err)
	t.Logf("MarshalSized(s) -> %X", b)

	var s2 stringWrapper
	var n int64
	_, err = cdc.UnmarshalSizedReader(bytes.NewBuffer(b), &s2, int64(len(b)-1))
	assert.NotNil(t, err, "insufficient limit should lead to failure")
	n, err = cdc.UnmarshalSizedReader(bytes.NewBuffer(b), &s2, int64(len(b)))
	assert.Nil(t, err, "sufficient limit should not cause failure")
	assert.Equal(t, s1, s2)
	frameLengthBytes, msgLengthBytes, embedOverhead := 1, 1, 1
	assert.Equal(t, frameLengthBytes+msgLengthBytes+embedOverhead+len(s1.S), int(n))
}

func TestUnmarshalReaderTooLong(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	type SimpleStruct struct {
		String string
		Bytes  []byte
		Time   time.Time
	}

	s := SimpleStruct{
		String: "hello",
		Bytes:  []byte("goodbye"),
		Time:   time.Now().UTC().Truncate(time.Millisecond), // strip monotonic and timezone.
	}

	b, err := cdc.MarshalSized(s)
	assert.Nil(t, err)
	t.Logf("MarshalSized(s) -> %X", b)

	var s2 SimpleStruct
	_, err = cdc.UnmarshalSizedReader(bytes.NewBuffer(b), &s2, 1) // 1 byte limit is ridiculous.
	assert.NotNil(t, err)
}

func TestUnmarshalBufferedWritesReads(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	buf := bytes.NewBuffer(nil)

	// Write 3 times.
	s1 := stringWrapper{"foo"}
	_, err := cdc.MarshalSizedWriter(buf, s1)
	assert.Nil(t, err)
	_, err = cdc.MarshalSizedWriter(buf, s1)
	assert.Nil(t, err)
	_, err = cdc.MarshalSizedWriter(buf, s1)
	assert.Nil(t, err)

	// Read 3 times.
	s2 := stringWrapper{}
	_, err = cdc.UnmarshalSizedReader(buf, &s2, 0)
	assert.Nil(t, err)
	assert.Equal(t, s1, s2)
	_, err = cdc.UnmarshalSizedReader(buf, &s2, 0)
	assert.Nil(t, err)
	assert.Equal(t, s1, s2)
	_, err = cdc.UnmarshalSizedReader(buf, &s2, 0)
	assert.Nil(t, err)
	assert.Equal(t, s1, s2)

	// Reading 4th time fails.
	_, err = cdc.UnmarshalSizedReader(buf, &s2, 0)
	assert.NotNil(t, err)
}

func TestBoolPointers(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	type SimpleStruct struct {
		BoolPtrTrue  *bool
		BoolPtrFalse *bool
	}

	ttrue := true
	ffalse := false

	s := SimpleStruct{
		BoolPtrTrue:  &ttrue,
		BoolPtrFalse: &ffalse,
	}

	b, err := cdc.Marshal(s)
	assert.NoError(t, err)

	var s2 SimpleStruct
	err = cdc.Unmarshal(b, &s2)

	assert.NoError(t, err)
	assert.NotNil(t, s2.BoolPtrTrue)
	assert.NotNil(t, s2.BoolPtrFalse)
}
