// Package crossrealm_f provides a collection realm for testing cross-realm
// ownership scenarios. It uses a nested structure so that intermediate objects
// in the ownership chain have RefCount == 1.
package crossrealm_f

type Entry struct {
	Key   string
	Value int
}

var entries []*Entry

func Add(cur realm, e *Entry) {
	entries = append(entries, e)
}

func Remove(cur realm, key string) *Entry {
	for i, e := range entries {
		if e.Key == key {
			entries = append(entries[:i], entries[i+1:]...)
			return e
		}
	}
	return nil
}

func Get(key string) *Entry {
	for _, e := range entries {
		if e.Key == key {
			return e
		}
	}
	return nil
}

func Len() int {
	return len(entries)
}
