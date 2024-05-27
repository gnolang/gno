package gno

import "sync"

type remoteApps struct {
	sync.RWMutex
	apps map[string]struct{}
}

var RemoteApps = remoteApps{apps: make(map[string]struct{})}

func (r *remoteApps) Add(app string) {
	r.Lock()
	r.apps[app] = struct{}{}
	r.Unlock()
}

func (r *remoteApps) Has(app string) bool {
	r.RLock()
	_, ok := r.apps[app]
	r.RUnlock()
	return ok
}
