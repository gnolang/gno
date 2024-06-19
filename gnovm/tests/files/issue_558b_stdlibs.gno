package main

import (
	"fmt"
	"io"
	"io"
	"log"
	"strings"
)

type readAutoCloser struct {
	r io.ReadCloser
}

func (a readAutoCloser) Read(b []byte) (n int, err error) {
	if a.r == nil {
		return 0, io.EOF
	}
	n, err = a.r.Read(b)
	if err == io.EOF {
		a.Close()
	}
	return n, err
}

func (a readAutoCloser) Close() error {
	if a.r == nil {
		return nil
	}
	return a.r.(io.Closer).Close()
}

type pipe struct {
	Reader readAutoCloser
}

func newReadAutoCloser(r io.Reader) readAutoCloser {
	if _, ok := r.(io.Closer); !ok {
		return readAutoCloser{io.NopCloser(r)}
	}
	return readAutoCloser{r.(io.ReadCloser)}
}

type Reader interface {
	Read(b []byte) (n int, err error)
}

func ReadAll(r Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}

func main() {
	p := &pipe{}
	p.Reader = newReadAutoCloser(strings.NewReader("test"))
	b, err := ReadAll(p.Reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

// Output:
// test
