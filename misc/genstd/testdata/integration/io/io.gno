package io

type Writer interface {
	Write([]byte) (int, error)
}
