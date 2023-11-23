package amino

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeFieldNumberAndTyp3_1(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := encodeFieldNumberAndTyp3(buf, 1, Typ3ByteLength)
	assert.Nil(t, err)
	assert.Equal(t, []byte{byte(0x01<<3 | Typ3ByteLength)}, buf.Bytes())
}

func TestEncodeFieldNumberAndTyp3_2(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := encodeFieldNumberAndTyp3(buf, 2, Typ3ByteLength)
	assert.Nil(t, err)
	assert.Equal(t, []byte{byte(0x02<<3 | Typ3ByteLength)}, buf.Bytes())
}
