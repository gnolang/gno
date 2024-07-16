package bytes

import "io"

type Buffer struct{ b []byte }

func (b *Buffer) Write(buf []byte) (int, error) {
	b.b = append(b.b, buf...)
	return len(buf), nil
}

var _ io.Writer = (*Buffer)(nil)

func myNative() int // injected
