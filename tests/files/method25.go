package main

import (
	"fmt"
	"sync"
)

func (p Pool) Get() *Buffer { return &Buffer{} }

func NewPool() Pool { return Pool{} }

type Buffer struct {
	Bs   []byte
	Pool Pool
}

type Pool struct {
	P *sync.Pool
}

var (
	_pool = NewPool()
	Get   = _pool.Get
)

func main() {
	fmt.Println(_pool)
	fmt.Println(Get())
}

// Output:
// {<nil>}
// &{[] {<nil>}}
