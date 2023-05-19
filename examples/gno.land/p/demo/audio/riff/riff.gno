package riff

import "io"

type Writer struct {
	io.Writer
}

func NewWriter(w io.Writer, fileType []byte, fileSize uint32) (w2 *Writer, err error) {
	w2 = &Writer{w}
	_, err = w2.Write([]byte("RIFF"))
	if err != nil {
		return
	}
	_, err = w2.WriteUint32(fileSize)
	if err != nil {
		return
	}
	_, err = w2.Write(fileType)
	if err != nil {
		return
	}
	return
}

func (w *Writer) WriteChunk(chunkID []byte, chunkSize uint32) (n int, err error) {
	n1, err := w.Write(chunkID)
	n = n1
	if err != nil {
		return
	}

	n2, err := w.WriteUint32(chunkSize)
	n += n2
	if err != nil {
		return
	}

	return
}

func (w *Writer) WriteUint16(v uint16) (n int, err error) {
	b := make([]byte, 2)
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	n, err = w.Write(b)
	return
}

func (w *Writer) WriteUint32(v uint32) (n int, err error) {
	b := make([]byte, 4)
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	n, err = w.Write(b)
	return
}

func Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[0]) | uint16(b[1])<<8
}
