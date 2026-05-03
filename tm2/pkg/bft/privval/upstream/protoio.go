package upstream

// protoio.go: tiny varint-length-prefixed reader/writer for proto.Message
// values over an io.Reader / io.Writer.
//
// CometBFT uses cometbft/libs/protoio for this. tm2 doesn't have an
// equivalent so we implement it locally — same wire shape: a varint
// length prefix followed by the message bytes. Compatible with the
// privval socket protocol upstream tmkms expects.
//
// Wire format per message:
//   <varint length><N bytes proto-encoded message>

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

// MaxRemoteSignerMsgSize bounds a single privval message. Mirrors
// cometbft/privval/signer_endpoint.go::maxRemoteSignerMsgSize. Larger
// messages are refused before allocation to prevent memory amplification.
const MaxRemoteSignerMsgSize = 1024 * 10

// DelimitedReader reads varint-length-prefixed proto.Message values.
type DelimitedReader struct {
	r          *bufio.Reader
	maxMsgSize int
}

// NewDelimitedReader wraps r with a buffered reader and enforces the given
// per-message size cap. Use MaxRemoteSignerMsgSize for the privval path.
func NewDelimitedReader(r io.Reader, maxMsgSize int) *DelimitedReader {
	return &DelimitedReader{r: bufio.NewReader(r), maxMsgSize: maxMsgSize}
}

// ReadMsg reads one message and decodes it into msg. Returns the number of
// bytes consumed (length-prefix + payload), or an error.
func (dr *DelimitedReader) ReadMsg(msg proto.Message) (int, error) {
	length64, err := binary.ReadUvarint(dr.r)
	if err != nil {
		return 0, err
	}
	if length64 > uint64(dr.maxMsgSize) {
		return 0, fmt.Errorf("upstream/protoio: message length %d exceeds cap %d", length64, dr.maxMsgSize)
	}
	length := int(length64)
	buf := make([]byte, length)
	if _, err := io.ReadFull(dr.r, buf); err != nil {
		return 0, fmt.Errorf("upstream/protoio: short read: %w", err)
	}
	if err := proto.Unmarshal(buf, msg); err != nil {
		return 0, fmt.Errorf("upstream/protoio: unmarshal: %w", err)
	}
	// length of the varint prefix on disk: re-derive from the value to
	// return total bytes consumed.
	return varintLen(length64) + length, nil
}

// DelimitedWriter writes varint-length-prefixed proto.Message values.
type DelimitedWriter struct {
	w io.Writer
}

// NewDelimitedWriter wraps w. The caller is responsible for buffering /
// flushing if desired; this writer makes one Write per message.
func NewDelimitedWriter(w io.Writer) *DelimitedWriter {
	return &DelimitedWriter{w: w}
}

// WriteMsg encodes msg, prepends a varint length, and emits the result.
// Returns the number of bytes written.
func (dw *DelimitedWriter) WriteMsg(msg proto.Message) (int, error) {
	if msg == nil {
		return 0, errors.New("upstream/protoio: nil message")
	}
	body, err := proto.Marshal(msg)
	if err != nil {
		return 0, fmt.Errorf("upstream/protoio: marshal: %w", err)
	}
	var prefix [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(prefix[:], uint64(len(body)))
	out := make([]byte, n+len(body))
	copy(out, prefix[:n])
	copy(out[n:], body)
	return dw.w.Write(out)
}

// varintLen returns the byte length of the unsigned varint encoding of v.
func varintLen(v uint64) int {
	n := 1
	for v >= 0x80 {
		v >>= 7
		n++
	}
	return n
}
