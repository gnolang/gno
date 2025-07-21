package main

import (
	"sync"
)

// pathManager manages a set of unique paths.
type pathManager struct {
	paths map[string]struct{}
	mu    sync.RWMutex
}

func newPathManager() *pathManager {
	return &pathManager{
		paths: make(map[string]struct{}),
	}
}

// Save add one path to the PathManager. If a path already exists, it is not added again.
func (p *pathManager) Save(path string) (exist bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exist = p.paths[path]; !exist {
		p.paths[path] = struct{}{}
	}
	return exist
}

func (p *pathManager) List() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	paths := make([]string, 0, len(p.paths))
	for path := range p.paths {
		paths = append(paths, path)
	}

	return paths
}

func (p *pathManager) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paths = make(map[string]struct{})
}
