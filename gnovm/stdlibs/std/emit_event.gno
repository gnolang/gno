package std

// Emit is a function that constructs a GnoEvent with a specified type and attributes.
// It then forwards this event to the event logger. Each emitted event carries metadata
// such as the timestamp of emission, the event type, and the provided attributes.
//
// The function takes arguments in pairs. The first element of each pair is considered
// the attribute key, and the second element is the corresponding value.
//
// The event is dispatched to the EventLogger, which resides in the tm2/pkg/sdk/events.go file.
//
// For more details about the GnoEvent data structure, refer to its definition in the emit_event.go file.
func Emit(typ string, attrs ...string) { emit(typ, attrs) }
func emit(typ string, attrs []string)
