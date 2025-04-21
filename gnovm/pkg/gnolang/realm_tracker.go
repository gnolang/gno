package gnolang

import (
	"sync"
)

// RealmTracker tracks realms that have been used during tests
// so they can be reset between test functions
type RealmTracker struct {
	mu         sync.Mutex
	realmPaths map[string]struct{}
}

func NewRealmTracker() *RealmTracker {
	return &RealmTracker{
		realmPaths: make(map[string]struct{}),
	}
}

func (rt *RealmTracker) TrackRealm(pkgPath string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.realmPaths[pkgPath] = struct{}{}
}

func (rt *RealmTracker) GetTrackedRealms() []string {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	result := make([]string, 0, len(rt.realmPaths))
	for path := range rt.realmPaths {
		result = append(result, path)
	}
	return result
}

// ResetMachineState should be used to provide an API to reset machine context and realm state between test functions.
func ResetMachineState(m *Machine, store Store, tracker *RealmTracker) error {
	m.ResetContext()
	return resetRealmState(store, tracker)
}

func resetRealmState(store Store, tracker *RealmTracker) error {
	realmPaths := tracker.GetTrackedRealms()

	for _, path := range realmPaths {
		cleanRealm := NewRealm(path)
		store.SetPackageRealm(cleanRealm)
	}

	store.ClearObjectCache()

	return nil
}

type InterceptStore struct {
	Store
	tracker *RealmTracker
}

func NewInterceptStore(store Store, tracker *RealmTracker) *InterceptStore {
	return &InterceptStore{
		Store:   store,
		tracker: tracker,
	}
}

func (is *InterceptStore) GetPackageRealm(pkgPath string) *Realm {
	is.tracker.TrackRealm(pkgPath)
	return is.Store.GetPackageRealm(pkgPath)
}

func (is *InterceptStore) SetPackageRealm(realm *Realm) {
	is.tracker.TrackRealm(realm.Path)
	is.Store.SetPackageRealm(realm)
}

type TestHelper struct {
	Machine *Machine
	Store   Store
	tracker *RealmTracker
}

func NewTestHelper(m *Machine, store Store) *TestHelper {
	tracker := NewRealmTracker()

	interceptStore := NewInterceptStore(store, tracker)
	m.Store = interceptStore

	return &TestHelper{
		Machine: m,
		Store:   interceptStore,
		tracker: tracker,
	}
}

func (t *TestHelper) ResetState() error {
	return ResetMachineState(t.Machine, t.Store, t.tracker)
}
