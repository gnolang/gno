package std

// Emit creates a GnoEvent (event) with the given type and attributes
// and passes the event to the event logger. emitted events contain a
// some metadata, such as the time of emission, the event type,
// and the attributes.
func Emit(typ string, attrs []string) { emit(typ, attrs) }
func emit(typ string, attrs []string)
