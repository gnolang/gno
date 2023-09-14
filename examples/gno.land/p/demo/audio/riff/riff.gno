// riff implements a simple writer for generating the 44-byte header
// inherit in .wav files. The specification is standard and this
// implementation follows the standard as specified here:
// https://www.mmsp.ece.mcgill.ca/Documents/AudioFormats/WAVE/WAVE.html
package riff

import "io"

// Writer implements an i/o writer for the header
type Writer struct {
	io.Writer
}

// NewWriter returns a new Writer given the io.Writer, the filetype as a byte slice,
// and the file size as an unsigned 32-bit integer. It writes the RIFF header for
// a .wav file to the provided io.Writer.
//
// Parameters:
//
//	w: The io.Writer where the header will be written.
//	fileType: The file type as a 4-byte slice (e.g., []byte("WAVE")).
//	fileSize: The size of the file in bytes.
//
// Returns:
//
//	w2: A pointer to the created Writer.
//	err: Any error that occurred during the write operation, if applicable.
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

// WriteChunk writes a new chunk to the current .wav file. It writes the chunk ID
// as a byte slice and the chunk size as an unsigned 32-bit integer in little-endian format.
//
// Parameters:
//
//	chunkID: The chunk ID as a byte slice.
//	chunkSize: The size of the chunk in bytes.
//
// Returns:
//
//	n: The number of bytes written to the underlying Writer.
//	err: Any error that occurred during the write operation, if applicable.
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

// WriteUint16 writes an unsigned 16-bit integer by first converting it to bytes
// in little-endian format and then writes the bytes to the underlying Writer.
//
// Parameters:
//
//	v: The unsigned 16-bit integer value to be written.
//
// Returns:
//
//	n: The number of bytes written to the underlying Writer.
//	err: Any error that occurred during the write operation, if applicable.
func (w *Writer) WriteUint16(v uint16) (n int, err error) {
	b := make([]byte, 2)
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	n, err = w.Write(b)
	return
}

// WriteUint32 writes an unsigned 32-bit integer by first converting it to
// bytes in little-endian format and then writes the bytes to the underlying Writer.
//
// Parameters:
//
//	v: The unsigned 32-bit integer value to be written.
//
// Returns:
//
//	n: The number of bytes written to the underlying Writer.
//	err: Any error that occurred during the write operation, if applicable.
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

// Uint32 converts a slice of bytes to an unsigned 32-bit integer,
// in little-endian format.
//
// The bytes is the slice of bytes to be converted.
func Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// Uint16 converts a slice of bytes to an unsigned 16-bit integer,
// in little-endian format.
//
// The bytes is the slice of bytes to be converted.
func Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[0]) | uint16(b[1])<<8
}
