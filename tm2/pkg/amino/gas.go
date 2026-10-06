package amino

// Gas constants for amino serialization compute cost.
// Calibrated from Binary2 (genproto2) benchmarks:
// ~2.8 ns/byte + 427 ns flat (see gnovm/adr/STORAGE_CHARGING_AMINO_HEURISTIC.png).
// The per-byte slope is used; the flat component is small relative to
// typical serialized sizes. Same slope assumed for decode pending
// separate benchmarks. These constants assume amino2 (Binary2).
const (
	GasEncodePerByte int64 = 3 // ~2.8 ns/byte amino marshal
	GasDecodePerByte int64 = 3 // ~2.8 ns/byte amino unmarshal (assumed)
)
