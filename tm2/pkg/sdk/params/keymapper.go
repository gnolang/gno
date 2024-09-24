package params

// KeyMapper is used to map one key string to another.
type KeyMapper interface {
	// Map does a transformation on an input key to produce the key
	// appropriate for accessing a param keeper's storage instance.
	Map(key string) string
}
