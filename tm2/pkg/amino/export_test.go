package amino

// WithoutInterfaceSetters runs f with no interface setter registered, so
// UnmarshalAnyBinary2 assigns through reflect.Value.Set as before.
func WithoutInterfaceSetters(f func()) {
	saved := map[any]any{}
	interfaceSetters.Range(func(k, v any) bool {
		saved[k] = v
		interfaceSetters.Delete(k)
		return true
	})
	defer func() {
		for k, v := range saved {
			interfaceSetters.Store(k, v)
		}
	}()
	f()
}
