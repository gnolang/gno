package main

import (
	"sync"
)

func NewPool() Pool { return Pool{} }

type Pool struct {
	p *sync.Pool
}

var _pool = NewPool()

func main() {
	println(_pool)
}

// Output:
// struct{(gonative{<nil>} gonative{*sync.Pool})}
